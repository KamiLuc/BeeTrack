import 'dart:convert';
import 'dart:typed_data';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/assistant/view/assistant_screen.dart';
import 'package:app/l10n/app_localizations.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

class _StreamingAdapter implements HttpClientAdapter {
  List<int> _utf8(String s) => utf8.encode(s);

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (!options.path.contains('/assistant/messages')) {
      return ResponseBody.fromString(
        jsonEncode({'count': 0}),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    final sse = 'event: delta\ndata: {"text": "Hi there"}\n\n'
        'event: done\ndata: {}\n\n';
    return ResponseBody(
      Stream.value(Uint8List.fromList(_utf8(sse))),
      200,
      headers: {
        Headers.contentTypeHeader: ['text/event-stream'],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient =
      ApiClient(storage: TokenStorage(prefs), baseUrl: 'https://api.test');
  apiClient.dio.httpClientAdapter = _StreamingAdapter();
  return apiClient;
}

/// Serves the conversation-history endpoints, tracking how many times the
/// list is fetched so the panel's refresh behavior can be verified.
class _HistoryAdapter implements HttpClientAdapter {
  _HistoryAdapter({this.conversationMessageCount = 2});

  final int conversationMessageCount;
  int conversationsRequestCount = 0;
  bool deleted = false;

  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (options.path == '/api/v1/assistant/conversations' && options.method == 'GET') {
      conversationsRequestCount++;
      return _json({
        'conversations': deleted
            ? []
            : [
                {
                  'id': 1,
                  'created_at': '2026-01-01T00:00:00Z',
                  'preview': 'how many hives?',
                  'message_count': conversationMessageCount,
                },
              ],
      });
    }
    if (options.path == '/api/v1/assistant/conversations/1/messages' && options.method == 'GET') {
      return _json({
        'conversation_id': 1,
        'messages': [
          for (var i = 0; i < conversationMessageCount; i++)
            {
              'role': i.isEven ? 'user' : 'assistant',
              'content': 'message $i',
              'created_at': '2026-01-01T00:00:00Z',
            },
        ],
      });
    }
    if (options.path == '/api/v1/assistant/conversations/1' && options.method == 'DELETE') {
      deleted = true;
      return ResponseBody.fromString('', 204);
    }
    return _json({'count': 0});
  }

  ResponseBody _json(Map<String, dynamic> data) => ResponseBody.fromString(
        jsonEncode(data),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
}

Future<ApiClient> _fakeApiClientWithHistory(_HistoryAdapter adapter) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient =
      ApiClient(storage: TokenStorage(prefs), baseUrl: 'https://api.test');
  apiClient.dio.httpClientAdapter = adapter;
  return apiClient;
}

Widget _wrap(ApiClient apiClient) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: RepositoryProvider<ApiClient>.value(
        value: apiClient,
        child: AssistantScreen(onSelectSection: (_) {}),
      ),
    );

void main() {
  testWidgets('shows the empty state when there are no messages',
      (tester) async {
    final apiClient = await _fakeApiClient();
    await tester.pumpWidget(_wrap(apiClient));
    await tester.pumpAndSettle();

    expect(
      find.text('Ask me anything related to beekeeping.'),
      findsOneWidget,
    );
  });

  testWidgets('sending a message renders the user bubble and streamed reply',
      (tester) async {
    final apiClient = await _fakeApiClient();
    await tester.pumpWidget(_wrap(apiClient));
    await tester.pumpAndSettle();

    await tester.enterText(find.byType(TextField), 'Hello bees');
    await tester.pump();
    await tester.tap(find.widgetWithIcon(IconButton, Icons.send));
    await tester.pumpAndSettle();

    expect(find.text('Hello bees'), findsOneWidget);
    expect(find.text('Hi there'), findsOneWidget);
  });

  testWidgets('the history panel is shown by default with the loaded conversation list',
      (tester) async {
    final adapter = _HistoryAdapter();
    final apiClient = await _fakeApiClientWithHistory(adapter);
    await tester.pumpWidget(_wrap(apiClient));
    await tester.pumpAndSettle();

    expect(find.text('how many hives?'), findsOneWidget);
    expect(find.text('New chat'), findsOneWidget);
  });

  testWidgets('the toggle button hides and re-shows the history panel',
      (tester) async {
    final adapter = _HistoryAdapter();
    final apiClient = await _fakeApiClientWithHistory(adapter);
    await tester.pumpWidget(_wrap(apiClient));
    await tester.pumpAndSettle();
    expect(find.text('how many hives?'), findsOneWidget);

    await tester.tap(find.byTooltip('Show/hide chat history'));
    await tester.pumpAndSettle();
    expect(find.text('how many hives?'), findsNothing);
    expect(find.text('New chat'), findsNothing);

    await tester.tap(find.byTooltip('Show/hide chat history'));
    await tester.pumpAndSettle();
    expect(find.text('how many hives?'), findsOneWidget);
  });

  testWidgets(
    'on a narrow screen the history panel stacks below the chat instead of beside it, without overflowing',
    (tester) async {
      tester.view.physicalSize = const Size(375, 812);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final adapter = _HistoryAdapter();
      final apiClient = await _fakeApiClientWithHistory(adapter);
      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text('how many hives?'), findsOneWidget);
      expect(find.text('Ask me anything related to beekeeping.'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('selecting a conversation loads it and refreshes the list',
      (tester) async {
    final adapter = _HistoryAdapter();
    final apiClient = await _fakeApiClientWithHistory(adapter);
    await tester.pumpWidget(_wrap(apiClient));
    await tester.pumpAndSettle();
    final requestsBeforeSelect = adapter.conversationsRequestCount;

    await tester.tap(find.text('how many hives?'));
    for (var i = 0; i < 20; i++) {
      await tester.pump(const Duration(milliseconds: 100));
    }

    expect(find.text('message 0'), findsOneWidget);
    expect(adapter.conversationsRequestCount, greaterThan(requestsBeforeSelect));
  });

  testWidgets('tapping "New chat" starts a fresh conversation',
      (tester) async {
    final adapter = _HistoryAdapter(conversationMessageCount: 10);
    final apiClient = await _fakeApiClientWithHistory(adapter);
    await tester.pumpWidget(_wrap(apiClient));
    await tester.pumpAndSettle();

    await tester.tap(find.text('how many hives?'));
    for (var i = 0; i < 20; i++) {
      await tester.pump(const Duration(milliseconds: 100));
    }

    expect(
      find.text(
        'This conversation has reached its message limit. Start a new chat to keep going.',
      ),
      findsOneWidget,
    );

    await tester.tap(find.text('New chat'));
    for (var i = 0; i < 20; i++) {
      await tester.pump(const Duration(milliseconds: 100));
    }

    expect(
      find.text(
        'This conversation has reached its message limit. Start a new chat to keep going.',
      ),
      findsNothing,
    );
  });

  testWidgets(
    'selecting a conversation at the message limit disables the input and shows the limit banner',
    (tester) async {
      final adapter = _HistoryAdapter(conversationMessageCount: 10);
      final apiClient = await _fakeApiClientWithHistory(adapter);
      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      await tester.tap(find.text('how many hives?'));
      for (var i = 0; i < 20; i++) {
        await tester.pump(const Duration(milliseconds: 100));
      }

      expect(
        find.text(
          'This conversation has reached its message limit. Start a new chat to keep going.',
        ),
        findsOneWidget,
      );
      final textField = tester.widget<TextField>(find.byType(TextField));
      expect(textField.enabled, isFalse);
      expect(find.byIcon(Icons.send), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsNothing);
    },
  );

  testWidgets('deleting a conversation via the confirmation dialog removes it from the list',
      (tester) async {
    final adapter = _HistoryAdapter();
    final apiClient = await _fakeApiClientWithHistory(adapter);
    await tester.pumpWidget(_wrap(apiClient));
    await tester.pumpAndSettle();
    expect(find.text('how many hives?'), findsOneWidget);

    await tester.tap(find.byIcon(Icons.delete_outline));
    await tester.pumpAndSettle();
    expect(find.text('Delete conversation?'), findsOneWidget);

    await tester.tap(find.text('Delete').last);
    await tester.pumpAndSettle();

    expect(adapter.deleted, isTrue);
    expect(find.text('how many hives?'), findsNothing);
  });
}
