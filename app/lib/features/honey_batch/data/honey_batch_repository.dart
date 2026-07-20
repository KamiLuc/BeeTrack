import 'package:dio/dio.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'honey_batch_model.dart';
import 'processing_method.dart';

class HoneyBatchRepository {
  final ApiClient _api;

  HoneyBatchRepository({required ApiClient api}) : _api = api;

  Future<({List<HoneyBatchModel> items, int total})> listBatches({
    int limit = 20,
    int offset = 0,
  }) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/honey-batches',
        queryParameters: {'limit': limit, 'offset': offset},
      );
      final body = response.data as Map<String, dynamic>;
      final items = (body['items'] as List<dynamic>)
          .map((e) => HoneyBatchModel.fromJson(e as Map<String, dynamic>))
          .toList();
      final total = body['total'] as int;
      return (items: items, total: total);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<HoneyBatchModel> getBatch(int id) async {
    try {
      final response = await _api.dio.get('/api/v1/honey-batches/$id');
      return HoneyBatchModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<HoneyBatchModel> createBatch({
    required DateTime gatheringDate,
    required int amountGrams,
    required ProcessingMethod processingMethod,
    required String honeyType,
    List<int>? pdfBytes,
    String? pdfFilename,
    bool requestCertification = false,
  }) async {
    try {
      final formData = FormData.fromMap({
        'gathering_date': DateFormat('yyyy-MM-dd').format(gatheringDate),
        'amount_grams': amountGrams.toString(),
        'processing_method': processingMethod.toJson(),
        'honey_type': honeyType,
        'request_certification': requestCertification.toString(),
        if (pdfBytes != null)
          'lab_pdf': MultipartFile.fromBytes(
            pdfBytes,
            filename: pdfFilename,
            contentType: DioMediaType.parse('application/pdf'),
          ),
      });
      final response = await _api.dio.post(
        '/api/v1/honey-batches',
        data: formData,
      );
      return HoneyBatchModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<HoneyBatchModel> updateHoneyType(int id, String honeyType) async {
    try {
      final response = await _api.dio.patch(
        '/api/v1/honey-batches/$id',
        data: {'honey_type': honeyType},
      );
      return HoneyBatchModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteBatch(int id) async {
    try {
      await _api.dio.delete('/api/v1/honey-batches/$id');
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<HoneyBatchModel> requestCertification(int id) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/honey-batches/$id/retry-certification',
      );
      return HoneyBatchModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<HoneyBatchModel> verifyByToken(String token) async {
    try {
      final response = await _api.dio.get('/api/v1/verify/$token');
      return HoneyBatchModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
