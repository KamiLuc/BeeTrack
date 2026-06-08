import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'harvest_model.dart';

class HarvestRepository {
  final ApiClient _api;

  HarvestRepository({required ApiClient api}) : _api = api;

  Future<({List<Harvest> items, int total})> listHarvests(
    int apiaryId,
    int hiveId, {
    int limit = 20,
    int offset = 0,
  }) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/harvests',
        queryParameters: {'limit': limit, 'offset': offset},
      );
      final body = response.data as Map<String, dynamic>;
      final items = (body['items'] as List<dynamic>)
          .map((e) => Harvest.fromJson(e as Map<String, dynamic>))
          .toList();
      final total = body['total'] as int;
      return (items: items, total: total);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Harvest> createHarvest({
    required int apiaryId,
    required int hiveId,
    required DateTime harvestedAt,
    required int frames,
    required int halfFrames,
    required double kilograms,
    required String notes,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/harvests',
        data: {
          'harvested_at': harvestedAt.toUtc().toIso8601String(),
          'frames': frames,
          'half_frames': halfFrames,
          'kilograms': kilograms,
          'notes': notes,
        },
      );
      return Harvest.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Harvest> updateHarvest({
    required int apiaryId,
    required int hiveId,
    required int harvestId,
    required DateTime harvestedAt,
    required int frames,
    required int halfFrames,
    required double kilograms,
    required String notes,
  }) async {
    try {
      final response = await _api.dio.patch(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/harvests/$harvestId',
        data: {
          'harvested_at': harvestedAt.toUtc().toIso8601String(),
          'frames': frames,
          'half_frames': halfFrames,
          'kilograms': kilograms,
          'notes': notes,
        },
      );
      return Harvest.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteHarvest({
    required int apiaryId,
    required int hiveId,
    required int harvestId,
  }) async {
    try {
      await _api.dio.delete(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/harvests/$harvestId',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
