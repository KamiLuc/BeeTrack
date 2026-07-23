import 'dart:convert';
import 'dart:typed_data';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/honey_batch/data/honey_batch_repository.dart';
import 'package:app/features/honey_batch/data/processing_method.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Records the last request made through it and returns a canned response.
class _RecordingAdapter implements HttpClientAdapter {
  RequestOptions? lastOptions;
  List<int>? lastBody;
  int statusCode = 200;
  Object? responseData;

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

Map<String, dynamic> _batchJson({int id = 1}) => {
      'id': id,
      'verification_token': 'tok-123',
      'gathering_date': '2024-05-01T00:00:00Z',
      'amount_grams': 2500,
      'processing_method': 'raw',
      'honey_type': 'Acacia',
      'pdf_file_hash': 'hash-abc',
      'created_at': '2024-05-02T08:00:00Z',
      'updated_at': '2024-05-03T09:00:00Z',
    };

void main() {
  late _RecordingAdapter adapter;
  late HoneyBatchRepository repository;

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();
    final storage = TokenStorage(prefs);
    final apiClient = ApiClient(storage: storage, baseUrl: 'https://api.test');
    adapter = _RecordingAdapter();
    apiClient.dio.httpClientAdapter = adapter;
    repository = HoneyBatchRepository(api: apiClient);
  });

  group('listBatches', () {
    test('sends GET with limit/offset query params and parses items', () async {
      adapter.responseData = {
        'items': [_batchJson(id: 1), _batchJson(id: 2)],
        'total': 2,
      };

      final result = await repository.listBatches(limit: 10, offset: 5);

      expect(adapter.lastOptions!.method, 'GET');
      expect(adapter.lastOptions!.path, '/api/v1/honey-batches');
      expect(adapter.lastOptions!.queryParameters, {'limit': 10, 'offset': 5});
      expect(result.total, 2);
      expect(result.items.map((e) => e.id), [1, 2]);
    });
  });

  group('getBatch', () {
    test('sends GET to /honey-batches/:id and parses batch', () async {
      adapter.responseData = _batchJson(id: 7);

      final batch = await repository.getBatch(7);

      expect(adapter.lastOptions!.method, 'GET');
      expect(adapter.lastOptions!.path, '/api/v1/honey-batches/7');
      expect(batch.id, 7);
    });
  });

  group('createBatch', () {
    test('sends multipart POST with expected fields and file', () async {
      adapter.responseData = _batchJson();

      final batch = await repository.createBatch(
        gatheringDate: DateTime(2024, 5, 1),
        amountGrams: 2500,
        processingMethod: ProcessingMethod.pasteurized,
        honeyType: 'Acacia',
        pdfBytes: [1, 2, 3],
        pdfFilename: 'lab.pdf',
        requestCertification: true,
      );

      expect(adapter.lastOptions!.method, 'POST');
      expect(adapter.lastOptions!.path, '/api/v1/honey-batches');

      final data = adapter.lastOptions!.data as FormData;
      final fields = {for (final f in data.fields) f.key: f.value};
      expect(fields['gathering_date'], '2024-05-01');
      expect(fields['amount_grams'], '2500');
      expect(fields['processing_method'], 'pasteurized');
      expect(fields['honey_type'], 'Acacia');
      expect(fields['request_certification'], 'true');
      expect(data.files.single.key, 'lab_pdf');
      expect(data.files.single.value.filename, 'lab.pdf');
      expect(batch.id, 1);
    });
  });

  group('updateBatch', () {
    test('sends multipart PATCH with the updated fields, no file', () async {
      adapter.responseData = _batchJson();

      await repository.updateBatch(
        id: 1,
        gatheringDate: DateTime(2024, 5, 1),
        amountGrams: 3000,
        processingMethod: ProcessingMethod.pasteurized,
        honeyType: 'Wildflower',
      );

      expect(adapter.lastOptions!.method, 'PATCH');
      expect(adapter.lastOptions!.path, '/api/v1/honey-batches/1');

      final data = adapter.lastOptions!.data as FormData;
      final fields = {for (final f in data.fields) f.key: f.value};
      expect(fields['gathering_date'], '2024-05-01');
      expect(fields['amount_grams'], '3000');
      expect(fields['processing_method'], 'pasteurized');
      expect(fields['honey_type'], 'Wildflower');
      expect(data.files, isEmpty);
    });

    test('sends the pdf file when replacing it', () async {
      adapter.responseData = _batchJson();

      await repository.updateBatch(
        id: 1,
        gatheringDate: DateTime(2024, 5, 1),
        amountGrams: 3000,
        processingMethod: ProcessingMethod.pasteurized,
        honeyType: 'Wildflower',
        pdfBytes: [1, 2, 3],
        pdfFilename: 'new-lab.pdf',
      );

      final data = adapter.lastOptions!.data as FormData;
      expect(data.files.single.key, 'lab_pdf');
      expect(data.files.single.value.filename, 'new-lab.pdf');
    });

    test('sends remove_pdf when removing without a replacement', () async {
      adapter.responseData = _batchJson();

      await repository.updateBatch(
        id: 1,
        gatheringDate: DateTime(2024, 5, 1),
        amountGrams: 3000,
        processingMethod: ProcessingMethod.pasteurized,
        honeyType: 'Wildflower',
        removePdf: true,
      );

      final data = adapter.lastOptions!.data as FormData;
      final fields = {for (final f in data.fields) f.key: f.value};
      expect(fields['remove_pdf'], 'true');
      expect(data.files, isEmpty);
    });
  });

  group('getPdfBytes', () {
    test('sends GET to /honey-batches/:id/pdf and returns raw bytes', () async {
      adapter.responseData = 'PDF-CONTENT';

      final bytes = await repository.getPdfBytes(9);

      expect(adapter.lastOptions!.method, 'GET');
      expect(adapter.lastOptions!.path, '/api/v1/honey-batches/9/pdf');
      expect(bytes, utf8.encode(jsonEncode('PDF-CONTENT')));
    });
  });

  group('deleteBatch', () {
    test('sends DELETE to /honey-batches/:id', () async {
      adapter.responseData = null;

      await repository.deleteBatch(3);

      expect(adapter.lastOptions!.method, 'DELETE');
      expect(adapter.lastOptions!.path, '/api/v1/honey-batches/3');
    });
  });

  group('requestCertification', () {
    test('sends POST to retry-certification endpoint', () async {
      adapter.responseData = _batchJson();

      await repository.requestCertification(1);

      expect(adapter.lastOptions!.method, 'POST');
      expect(
        adapter.lastOptions!.path,
        '/api/v1/honey-batches/1/retry-certification',
      );
    });
  });

  group('verifyByToken', () {
    test('sends GET to /verify/:token and parses batch', () async {
      adapter.responseData = _batchJson();

      await repository.verifyByToken('tok-123');

      expect(adapter.lastOptions!.method, 'GET');
      expect(adapter.lastOptions!.path, '/api/v1/verify/tok-123');
    });
  });
}
