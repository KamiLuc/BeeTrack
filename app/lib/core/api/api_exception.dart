import 'package:dio/dio.dart';

class ApiException implements Exception {
  final String code;
  final String message;

  const ApiException({required this.code, required this.message});

  factory ApiException.fromDioException(DioException e) {
    final data = e.response?.data;
    if (data is Map<String, dynamic>) {
      return ApiException(
        code: data['code'] as String? ?? 'UNKNOWN',
        message: data['message'] as String? ?? 'Unknown error',
      );
    }
    return const ApiException(code: 'NETWORK_ERROR', message: 'Network error');
  }

  @override
  String toString() => 'ApiException($code): $message';
}
