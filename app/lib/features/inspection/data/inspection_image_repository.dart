import 'package:dio/dio.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'inspection_image_model.dart';

class InspectionImageRepository {
  final ApiClient _api;

  InspectionImageRepository({required this._api});

  String imageUrl(int apiaryId, int hiveId, int inspectionId, int imageId) {
    return '${_api.baseUrl}/api/v1/apiaries/$apiaryId/hives/$hiveId/inspections/$inspectionId/images/$imageId/file';
  }

  Map<String, String> authHeaders() {
    final token = _api.accessToken;
    if (token == null) return {};
    return {'Authorization': 'Bearer $token'};
  }

  Future<List<InspectionImage>> listImages(
    int apiaryId,
    int hiveId,
    int inspectionId,
  ) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/inspections/$inspectionId/images',
      );
      final data = response.data as List<dynamic>;
      return data
          .map((e) => InspectionImage.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<InspectionImage> uploadImage(
    int apiaryId,
    int hiveId,
    int inspectionId,
    XFile file, {
    ProgressCallback? onSendProgress,
  }) async {
    try {
      final bytes = await file.readAsBytes();
      final mimeType = file.mimeType ?? _mimeFromPath(file.path);
      final formData = FormData.fromMap({
        'image': MultipartFile.fromBytes(
          bytes,
          filename: file.name,
          contentType: DioMediaType.parse(mimeType),
        ),
      });
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/inspections/$inspectionId/images',
        data: formData,
        onSendProgress: onSendProgress,
      );
      return InspectionImage.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteImage(
    int apiaryId,
    int hiveId,
    int inspectionId,
    int imageId,
  ) async {
    try {
      await _api.dio.delete(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/inspections/$inspectionId/images/$imageId',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  String _mimeFromPath(String path) {
    final ext = path.split('.').last.toLowerCase();
    return switch (ext) {
      'jpg' || 'jpeg' => 'image/jpeg',
      'png' => 'image/png',
      'webp' => 'image/webp',
      _ => 'image/jpeg',
    };
  }
}
