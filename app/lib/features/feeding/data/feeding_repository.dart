import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'feeding_model.dart';

class FeedingRepository {
  final ApiClient _api;

  FeedingRepository({required ApiClient api}) : _api = api;

  Future<List<String>> listFeedTypes() async {
    try {
      final response = await _api.dio.get('/api/v1/feed-types');
      return (response.data as List<dynamic>).cast<String>();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<List<String>> listAmounts() async {
    try {
      final response = await _api.dio.get('/api/v1/feed-amounts');
      return (response.data as List<dynamic>).cast<String>();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<({List<Feeding> items, int total})> listFeedings(
    int apiaryId,
    int hiveId, {
    int limit = 20,
    int offset = 0,
  }) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/feedings',
        queryParameters: {'limit': limit, 'offset': offset},
      );
      final body = response.data as Map<String, dynamic>;
      final items = (body['items'] as List<dynamic>)
          .map((e) => Feeding.fromJson(e as Map<String, dynamic>))
          .toList();
      final total = body['total'] as int;
      return (items: items, total: total);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Feeding> createFeeding({
    required int apiaryId,
    required int hiveId,
    required DateTime fedAt,
    required String feedType,
    required String amount,
    required String notes,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/feedings',
        data: {
          'fed_at': fedAt.toUtc().toIso8601String(),
          'feed_type': feedType,
          'amount': amount,
          'notes': notes,
        },
      );
      return Feeding.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Feeding> updateFeeding({
    required int apiaryId,
    required int hiveId,
    required int feedingId,
    required DateTime fedAt,
    required String feedType,
    required String amount,
    required String notes,
  }) async {
    try {
      final response = await _api.dio.patch(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/feedings/$feedingId',
        data: {
          'fed_at': fedAt.toUtc().toIso8601String(),
          'feed_type': feedType,
          'amount': amount,
          'notes': notes,
        },
      );
      return Feeding.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<int> bulkFeeding({
    required int apiaryId,
    List<int>? hiveIds,
    required DateTime fedAt,
    required String feedType,
    required String amount,
    required String notes,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/feedings/bulk',
        data: {
          if (hiveIds != null) 'hive_ids': hiveIds,
          'fed_at': fedAt.toUtc().toIso8601String(),
          'feed_type': feedType,
          'amount': amount,
          'notes': notes,
        },
      );
      return response.data['count'] as int;
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteFeeding({
    required int apiaryId,
    required int hiveId,
    required int feedingId,
  }) async {
    try {
      await _api.dio.delete(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/feedings/$feedingId',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
