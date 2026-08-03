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
}
