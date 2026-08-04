import 'dart:convert';
import 'dart:typed_data';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/api/api_exception.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/voice/data/voice_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Records the last request made through it and returns a canned response.
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

Map<String, dynamic> _recordingJson({int recordingId = 1, String status = 'pending'}) => {
      'recording_id': recordingId,
      'status': status,
      'transcript': null,
      'error_message': null,
      'created_at': null,
      'processed_at': null,
      'voice_actions': [],
    };

void main() {
  late _RecordingAdapter adapter;
  late VoiceRepository repository;

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();
    final storage = TokenStorage(prefs);
    final apiClient = ApiClient(storage: storage, baseUrl: 'https://api.test');
    adapter = _RecordingAdapter();
    apiClient.dio.httpClientAdapter = adapter;
    repository = VoiceRepository(api: apiClient);
  });

  group('uploadRecording', () {
    test('sends multipart POST and parses recording_id/status', () async {
      adapter.statusCode = 202;
      adapter.responseData = {'recording_id': 42, 'status': 'pending'};

      final recording = await repository.uploadRecording(
        3,
        [1, 2, 3],
        filename: 'voice.m4a',
        mimeType: 'audio/mp4',
      );

      expect(adapter.lastOptions!.method, 'POST');
      expect(adapter.lastOptions!.path, '/api/v1/apiaries/3/voice');
      final data = adapter.lastOptions!.data as FormData;
      expect(data.files.single.key, 'audio');
      expect(data.files.single.value.filename, 'voice.m4a');

      expect(recording.recordingId, 42);
      expect(recording.status, 'pending');
      expect(recording.transcript, isNull);
      expect(recording.errorMessage, isNull);
      expect(recording.createdAt, isNull);
      expect(recording.processedAt, isNull);
      expect(recording.voiceActions, isEmpty);
      expect(recording.localPath, isNull);
    });

    test('translates DioException into ApiException', () async {
      adapter.statusCode = 400;
      adapter.responseData = {'code': 'BAD_AUDIO', 'message': 'invalid audio'};

      await expectLater(
        () => repository.uploadRecording(
          3,
          [1, 2, 3],
          filename: 'voice.m4a',
          mimeType: 'audio/mp4',
        ),
        throwsA(
          isA<ApiException>()
              .having((e) => e.code, 'code', 'BAD_AUDIO')
              .having((e) => e.message, 'message', 'invalid audio'),
        ),
      );
    });
  });

  group('listRecordings', () {
    test('sends GET with limit/offset query params and parses items/total', () async {
      adapter.responseData = {
        'items': [_recordingJson(recordingId: 1), _recordingJson(recordingId: 2)],
        'total': 2,
      };

      final result = await repository.listRecordings(3, limit: 10, offset: 5);

      expect(adapter.lastOptions!.method, 'GET');
      expect(adapter.lastOptions!.path, '/api/v1/apiaries/3/voice-recordings');
      expect(adapter.lastOptions!.queryParameters, {'limit': 10, 'offset': 5});
      expect(result.total, 2);
      expect(result.items.map((e) => e.recordingId), [1, 2]);
    });

    test('translates DioException into ApiException', () async {
      adapter.statusCode = 500;
      adapter.responseData = {'code': 'SERVER_ERROR', 'message': 'oops'};

      await expectLater(
        () => repository.listRecordings(3),
        throwsA(
          isA<ApiException>()
              .having((e) => e.code, 'code', 'SERVER_ERROR')
              .having((e) => e.message, 'message', 'oops'),
        ),
      );
    });
  });

  group('cancelRecording', () {
    test('sends DELETE to the recording path', () async {
      adapter.statusCode = 204;
      adapter.responseData = null;

      await repository.cancelRecording(3, 42);

      expect(adapter.lastOptions!.method, 'DELETE');
      expect(adapter.lastOptions!.path, '/api/v1/apiaries/3/voice-recordings/42');
    });

    test('translates DioException into ApiException', () async {
      adapter.statusCode = 409;
      adapter.responseData = {
        'code': 'RECORDING_NOT_CANCELABLE',
        'message': 'recording is no longer cancelable',
      };

      await expectLater(
        () => repository.cancelRecording(3, 42),
        throwsA(
          isA<ApiException>()
              .having((e) => e.code, 'code', 'RECORDING_NOT_CANCELABLE')
              .having((e) => e.message, 'message', 'recording is no longer cancelable'),
        ),
      );
    });
  });

  group('updateActionArguments', () {
    test('sends PATCH with tool_arguments and parses the returned action', () async {
      adapter.responseData = {
        'id': 5,
        'sequence': 0,
        'hive_id': 7,
        'tool_name': 'log_feeding',
        'status': 'proposed',
        'result_type': null,
        'result_record_id': null,
        'error_message': null,
      };

      final action = await repository.updateActionArguments(
        3,
        42,
        5,
        {'feed_type': 'Sugar syrup', 'amount': '1L'},
      );

      expect(adapter.lastOptions!.method, 'PATCH');
      expect(
        adapter.lastOptions!.path,
        '/api/v1/apiaries/3/voice-recordings/42/actions/5',
      );
      expect(adapter.lastOptions!.data, {
        'tool_arguments': {'feed_type': 'Sugar syrup', 'amount': '1L'},
      });
      expect(action.id, 5);
      expect(action.toolName, 'log_feeding');
      expect(action.status, 'proposed');
    });

    test('translates DioException into ApiException', () async {
      adapter.statusCode = 409;
      adapter.responseData = {
        'code': 'ACTION_NOT_EDITABLE',
        'message': 'action is no longer proposed',
      };

      await expectLater(
        () => repository.updateActionArguments(3, 42, 5, {'amount': '1L'}),
        throwsA(
          isA<ApiException>()
              .having((e) => e.code, 'code', 'ACTION_NOT_EDITABLE')
              .having((e) => e.message, 'message', 'action is no longer proposed'),
        ),
      );
    });
  });
}
