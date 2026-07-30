import 'dart:convert';
import 'dart:typed_data';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/api/api_exception.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/assistant/data/assistant_chat_message.dart';
import 'package:app/features/assistant/data/assistant_event.dart';
import 'package:app/features/assistant/data/assistant_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Records the last request made through it and returns a canned JSON response.
class _RecordingAdapter implements HttpClientAdapter {
  RequestOptions? lastOptions;
  int statusCode = 200;
  Object? responseData;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    lastOptions = options;
    return ResponseBody.fromString(
      jsonEncode(responseData),
      statusCode,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

class _StreamingAdapter implements HttpClientAdapter {
  RequestOptions? lastOptions;
  List<int>? lastBody;
  List<List<int>> chunks = [];
  int statusCode = 200;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    lastOptions = options;
    lastBody = requestStream == null
        ? null
        : await requestStream.fold<List<int>>(
            <int>[], (prev, chunk) => prev..addAll(chunk));
    return ResponseBody(
      Stream.fromIterable(
        chunks.map((c) => Uint8List.fromList(c)),
      ),
      statusCode,
      headers: {
        Headers.contentTypeHeader: ['text/event-stream'],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

List<int> _utf8(String s) => utf8.encode(s);

void main() {
  late _StreamingAdapter adapter;
  late AssistantRepository repository;

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();
    final storage = TokenStorage(prefs);
    final apiClient = ApiClient(storage: storage, baseUrl: 'https://api.test');
    adapter = _StreamingAdapter();
    apiClient.dio.httpClientAdapter = adapter;
    repository = AssistantRepository(api: apiClient);
  });

  test('sends POST with message and conversation_id', () async {
    adapter.chunks = [
      _utf8('event: done\ndata: {}\n\n'),
    ];

    await repository
        .sendMessage(message: 'hello', conversationId: 42)
        .toList();

    expect(adapter.lastOptions!.method, 'POST');
    expect(adapter.lastOptions!.path, '/api/v1/assistant/messages');
    final body = jsonDecode(utf8.decode(adapter.lastBody!)) as Map;
    expect(body['message'], 'hello');
    expect(body['conversation_id'], 42);
  });

  test('omits conversation_id when null', () async {
    adapter.chunks = [
      _utf8('event: done\ndata: {}\n\n'),
    ];

    await repository.sendMessage(message: 'hi').toList();

    final body = jsonDecode(utf8.decode(adapter.lastBody!)) as Map;
    expect(body.containsKey('conversation_id'), isFalse);
  });

  test('parses a single event delivered in one chunk', () async {
    adapter.chunks = [
      _utf8('event: conversation\ndata: {"conversation_id": 7}\n\n'),
    ];

    final events = await repository.sendMessage(message: 'hi').toList();

    expect(events, hasLength(1));
    expect(events.single, isA<AssistantConversationEvent>());
    expect(
      (events.single as AssistantConversationEvent).conversationId,
      7,
    );
  });

  test('parses multiple events delivered in a single chunk', () async {
    adapter.chunks = [
      _utf8(
        'event: delta\ndata: {"text": "Hel"}\n\n'
        'event: delta\ndata: {"text": "lo"}\n\n'
        'event: done\ndata: {}\n\n',
      ),
    ];

    final events = await repository.sendMessage(message: 'hi').toList();

    expect(events, hasLength(3));
    expect((events[0] as AssistantDeltaEvent).text, 'Hel');
    expect((events[1] as AssistantDeltaEvent).text, 'lo');
    expect(events[2], isA<AssistantDoneEvent>());
  });

  test('reassembles a single event split across multiple chunks', () async {
    const raw = 'event: delta\ndata: {"text": "Hello"}\n\n';
    final bytes = _utf8(raw);
    final mid = bytes.length ~/ 2;
    adapter.chunks = [
      bytes.sublist(0, mid),
      bytes.sublist(mid),
    ];

    final events = await repository.sendMessage(message: 'hi').toList();

    expect(events, hasLength(1));
    expect((events.single as AssistantDeltaEvent).text, 'Hello');
  });

  test('parses error events', () async {
    adapter.chunks = [
      _utf8('event: error\ndata: {"message": "boom"}\n\n'),
    ];

    final events = await repository.sendMessage(message: 'hi').toList();

    expect(events.single, isA<AssistantErrorEvent>());
    expect((events.single as AssistantErrorEvent).message, 'boom');
  });

  test('throws ApiException decoded from an undecoded ResponseBody error',
      () async {
    adapter.statusCode = 400;
    adapter.chunks = [
      _utf8(jsonEncode({'code': 'BAD_REQUEST', 'message': 'nope'})),
    ];

    await expectLater(
      repository.sendMessage(message: 'hi').toList(),
      throwsA(
        isA<ApiException>()
            .having((e) => e.code, 'code', 'BAD_REQUEST')
            .having((e) => e.message, 'message', 'nope'),
      ),
    );
  });

  group('fetchConversations', () {
    test('GETs and parses the conversation list', () async {
      final recorder = _RecordingAdapter()
        ..responseData = {
          'conversations': [
            {
              'id': 1,
              'created_at': '2026-01-01T00:00:00Z',
              'preview': 'how many hives?',
              'message_count': 4,
            },
          ],
        };
      final prefs2 = await SharedPreferences.getInstance();
      final apiClient = ApiClient(
        storage: TokenStorage(prefs2),
        baseUrl: 'https://api.test',
      )..dio.httpClientAdapter = recorder;
      final repo = AssistantRepository(api: apiClient);

      final conversations = await repo.fetchConversations();

      expect(recorder.lastOptions!.method, 'GET');
      expect(recorder.lastOptions!.path, '/api/v1/assistant/conversations');
      expect(conversations, hasLength(1));
      expect(conversations.single.id, 1);
      expect(conversations.single.preview, 'how many hives?');
      expect(conversations.single.messageCount, 4);
    });
  });

  group('fetchMessages', () {
    test('GETs and parses the message history', () async {
      final recorder = _RecordingAdapter()
        ..responseData = {
          'conversation_id': 1,
          'messages': [
            {'role': 'user', 'content': 'hi', 'created_at': '2026-01-01T00:00:00Z'},
            {'role': 'assistant', 'content': 'hello', 'created_at': '2026-01-01T00:00:01Z'},
          ],
        };
      final prefs2 = await SharedPreferences.getInstance();
      final apiClient = ApiClient(
        storage: TokenStorage(prefs2),
        baseUrl: 'https://api.test',
      )..dio.httpClientAdapter = recorder;
      final repo = AssistantRepository(api: apiClient);

      final messages = await repo.fetchMessages(1);

      expect(recorder.lastOptions!.method, 'GET');
      expect(recorder.lastOptions!.path, '/api/v1/assistant/conversations/1/messages');
      expect(messages, hasLength(2));
      expect(messages[0].role, AssistantChatRole.user);
      expect(messages[0].text, 'hi');
      expect(messages[1].role, AssistantChatRole.assistant);
      expect(messages[1].text, 'hello');
    });
  });

  group('deleteConversation', () {
    test('sends DELETE to the conversation endpoint', () async {
      final recorder = _RecordingAdapter()..statusCode = 204;
      final prefs2 = await SharedPreferences.getInstance();
      final apiClient = ApiClient(
        storage: TokenStorage(prefs2),
        baseUrl: 'https://api.test',
      )..dio.httpClientAdapter = recorder;
      final repo = AssistantRepository(api: apiClient);

      await repo.deleteConversation(5);

      expect(recorder.lastOptions!.method, 'DELETE');
      expect(recorder.lastOptions!.path, '/api/v1/assistant/conversations/5');
    });
  });
}
