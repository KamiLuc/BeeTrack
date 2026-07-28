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
}
