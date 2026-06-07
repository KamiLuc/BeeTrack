import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'hive_model.dart';

class HiveRepository {
  final ApiClient _api;

  HiveRepository({required this._api});

  Future<Hive> getHive(int apiaryId, int hiveId) async {
    try {
      final response = await _api.dio
          .get('/api/v1/apiaries/$apiaryId/hives/$hiveId');
      return Hive.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<List<Hive>> listHives(int apiaryId) async {
    try {
      final response = await _api.dio.get('/api/v1/apiaries/$apiaryId/hives');
      final data = response.data as List<dynamic>;
      return data
          .map((e) => Hive.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> createHive({
    required int apiaryId,
    required String name,
    required String type,
    required bool active,
    required bool queenless,
    required bool readyForHarvest,
    required int gridRow,
    required int gridCol,
    int frames = 0,
  }) async {
    try {
      await _api.dio.post('/api/v1/apiaries/$apiaryId/hives', data: {
        'active': active,
        'frames': frames,
        'grid_col': gridCol,
        'grid_row': gridRow,
        'name': name,
        'queenless': queenless,
        'ready_for_harvest': readyForHarvest,
        'type': type,
      });
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> updateHive({
    required int apiaryId,
    required int hiveId,
    required String name,
    required String type,
    required bool active,
    required bool queenless,
    required bool readyForHarvest,
    int frames = 0,
  }) async {
    try {
      await _api.dio.patch('/api/v1/apiaries/$apiaryId/hives/$hiveId', data: {
        'active': active,
        'frames': frames,
        'name': name,
        'queenless': queenless,
        'ready_for_harvest': readyForHarvest,
        'type': type,
      });
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> addFrames({
    required int apiaryId,
    required int hiveId,
    required int delta,
  }) async {
    try {
      await _api.dio.patch(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/frames',
        data: {'delta': delta},
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteHive({
    required int apiaryId,
    required int hiveId,
  }) async {
    try {
      await _api.dio.delete('/api/v1/apiaries/$apiaryId/hives/$hiveId');
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> moveHive({
    required int apiaryId,
    required int hiveId,
    required int row,
    required int col,
  }) async {
    try {
      await _api.dio.patch(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/position',
        data: {'grid_row': row, 'grid_col': col},
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<HiveDisease> addDisease({
    required int apiaryId,
    required int hiveId,
    required String disease,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/diseases',
        data: {'disease': disease},
      );
      return HiveDisease.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> removeDisease({
    required int apiaryId,
    required int hiveId,
    required int diseaseId,
  }) async {
    try {
      await _api.dio.delete(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/diseases/$diseaseId',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Hive> changeApiary({
    required int apiaryId,
    required int hiveId,
    required int targetApiaryId,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/transfer',
        data: {'target_apiary_id': targetApiaryId},
      );
      return Hive.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
