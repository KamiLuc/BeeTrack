import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import '../../../core/storage/token_storage.dart';

class AuthRepository {
  final ApiClient _api;
  final TokenStorage _storage;

  AuthRepository({required this._api, required this._storage});

  Future<void> login({required String email, required String password}) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/auth/login',
        data: {'email': email, 'password': password},
      );
      await _storage.save(
        access: response.data['access_token'],
        refresh: response.data['refresh_token'],
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> register({
    required String email,
    required String name,
    required String password,
  }) async {
    try {
      await _api.dio.post(
        '/api/v1/auth/register',
        data: {'email': email, 'name': name, 'password': password},
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> logout() async {
    final refresh = _storage.refreshToken;
    if (refresh != null) {
      try {
        await _api.dio.post(
          '/api/v1/auth/logout',
          data: {'refresh_token': refresh},
        );
      } on DioException {
        // best-effort — clear tokens regardless
      }
    }
    await _storage.clear();
  }

  bool get isLoggedIn => _storage.accessToken != null;
}
