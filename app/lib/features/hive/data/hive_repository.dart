import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'hive_model.dart';

class HiveRepository {
  final ApiClient _api;

  HiveRepository({required this._api});

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
}
