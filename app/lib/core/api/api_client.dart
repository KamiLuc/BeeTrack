import 'package:dio/dio.dart';

import '../storage/token_storage.dart';

class ApiClient {
  final Dio _dio;
  final TokenStorage _storage;

  ApiClient({required this._storage, required String baseUrl})
      : _dio = Dio(BaseOptions(baseUrl: baseUrl)) {
    _dio.interceptors.add(_AuthInterceptor(_dio, _storage));
  }

  Dio get dio => _dio;

  String get baseUrl => _dio.options.baseUrl;

  String? get accessToken => _storage.accessToken;
}

class _AuthInterceptor extends Interceptor {
  final Dio _dio;
  final TokenStorage _storage;

  _AuthInterceptor(this._dio, this._storage);

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final token = _storage.accessToken;
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (err.response?.statusCode != 401) {
      handler.next(err);
      return;
    }

    final path = err.requestOptions.path;
    if (path.contains('/auth/')) {
      handler.next(err);
      return;
    }

    final refreshToken = _storage.refreshToken;
    if (refreshToken == null) {
      await _storage.clear();
      handler.next(err);
      return;
    }

    try {
      final response = await _dio.post(
        '/api/v1/auth/refresh',
        data: {'refresh_token': refreshToken},
        options: Options(extra: {'skipAuth': true}),
      );
      await _storage.save(
        access: response.data['access_token'],
        refresh: response.data['refresh_token'],
      );
      final retried = await _dio.fetch(err.requestOptions
        ..headers['Authorization'] =
            'Bearer ${_storage.accessToken}');
      handler.resolve(retried);
    } catch (_) {
      await _storage.clear();
      handler.next(err);
    }
  }
}
