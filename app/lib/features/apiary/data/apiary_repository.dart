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
}
