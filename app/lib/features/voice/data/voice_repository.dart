import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'voice_recording_model.dart';

class VoiceRepository {
  final ApiClient _api;

  VoiceRepository({required ApiClient api}) : _api = api;

  Future<VoiceRecording> uploadRecording(
    int apiaryId,
    List<int> bytes, {
    required String filename,
    required String mimeType,
  }) async {
    try {
      final formData = FormData.fromMap({
        'audio': MultipartFile.fromBytes(
          bytes,
          filename: filename,
          contentType: DioMediaType.parse(mimeType),
        ),
      });
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/voice',
        data: formData,
      );
      final body = response.data as Map<String, dynamic>;
      return VoiceRecording(
        recordingId: body['recording_id'] as int,
        status: body['status'] as String,
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<({List<VoiceRecording> items, int total})> listRecordings(
    int apiaryId, {
    int limit = 20,
    int offset = 0,
  }) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/apiaries/$apiaryId/voice-recordings',
        queryParameters: {'limit': limit, 'offset': offset},
      );
      final body = response.data as Map<String, dynamic>;
      final items = (body['items'] as List<dynamic>)
          .map((e) => VoiceRecording.fromJson(e as Map<String, dynamic>))
          .toList();
      return (items: items, total: body['total'] as int);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> cancelRecording(int apiaryId, int recordingId) async {
    try {
      await _api.dio.delete(
        '/api/v1/apiaries/$apiaryId/voice-recordings/$recordingId',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> rejectRecording(int apiaryId, int recordingId) async {
    try {
      await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/voice-recordings/$recordingId/reject',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> acceptRecording(int apiaryId, int recordingId) async {
    try {
      await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/voice-recordings/$recordingId/accept',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
