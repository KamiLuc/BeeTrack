import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'treatment_model.dart';

class TreatmentRepository {
  final ApiClient _api;

  TreatmentRepository({required ApiClient api}) : _api = api;

  Future<List<String>> listMedicines() async {
    try {
      final response = await _api.dio.get('/api/v1/medicines');
      return (response.data as List<dynamic>).cast<String>();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<({List<Treatment> items, int total})> listTreatments(
    int apiaryId,
    int hiveId, {
    int limit = 20,
    int offset = 0,
  }) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/treatments',
        queryParameters: {'limit': limit, 'offset': offset},
      );
      final body = response.data as Map<String, dynamic>;
      final items = (body['items'] as List<dynamic>)
          .map((e) => Treatment.fromJson(e as Map<String, dynamic>))
          .toList();
      final total = body['total'] as int;
      return (items: items, total: total);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Treatment> createTreatment({
    required int apiaryId,
    required int hiveId,
    required DateTime treatedAt,
    required String medicineName,
    required String dose,
    required String notes,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/treatments',
        data: {
          'treated_at': treatedAt.toUtc().toIso8601String(),
          'medicine_name': medicineName,
          'dose': dose,
          'notes': notes,
        },
      );
      return Treatment.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Treatment> updateTreatment({
    required int apiaryId,
    required int hiveId,
    required int treatmentId,
    required DateTime treatedAt,
    required String medicineName,
    required String dose,
    required String notes,
  }) async {
    try {
      final response = await _api.dio.patch(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/treatments/$treatmentId',
        data: {
          'treated_at': treatedAt.toUtc().toIso8601String(),
          'medicine_name': medicineName,
          'dose': dose,
          'notes': notes,
        },
      );
      return Treatment.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<int> bulkTreatment({
    required int apiaryId,
    required DateTime treatedAt,
    required String medicineName,
    required String dose,
    required String notes,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/treatments/bulk',
        data: {
          'treated_at': treatedAt.toUtc().toIso8601String(),
          'medicine_name': medicineName,
          'dose': dose,
          'notes': notes,
        },
      );
      return response.data['count'] as int;
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteTreatment({
    required int apiaryId,
    required int hiveId,
    required int treatmentId,
  }) async {
    try {
      await _api.dio.delete(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/treatments/$treatmentId',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
