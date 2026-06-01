import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'apiary_model.dart';

class ApiaryRepository {
  final ApiClient _api;

  ApiaryRepository({required this._api});

  Future<List<Apiary>> listApiaries() async {
    try {
      final response = await _api.dio.get('/api/v1/apiaries');
      final data = response.data as List<dynamic>;
      return data
          .map((e) => Apiary.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> createApiary({
    required String name,
    double? lat,
    double? lng,
    required int gridRows,
    required int gridCols,
  }) async {
    try {
      await _api.dio.post('/api/v1/apiaries', data: {
        'name': name,
        'lat': lat,
        'lng': lng,
        'grid_rows': gridRows,
        'grid_cols': gridCols,
      });
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> updateApiary({
    required int id,
    required String name,
    double? lat,
    double? lng,
    required int gridRows,
    required int gridCols,
  }) async {
    try {
      await _api.dio.patch('/api/v1/apiaries/$id', data: {
        'name': name,
        'lat': lat,
        'lng': lng,
        'grid_rows': gridRows,
        'grid_cols': gridCols,
      });
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteApiary(int id) async {
    try {
      await _api.dio.delete('/api/v1/apiaries/$id');
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
