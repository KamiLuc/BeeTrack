import 'package:dio/dio.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'listing_model.dart';

class ListingSearchResult {
  final List<Listing> items;
  final int total;

  const ListingSearchResult({required this.items, required this.total});
}

class ListingRepository {
  final ApiClient _api;

  ListingRepository({required this._api});

  Future<ListingSearchResult> searchListings({
    String? category,
    String? keyword,
    double? priceMin,
    double? priceMax,
    String? postedAfter,
    double? nearLat,
    double? nearLng,
    double? radiusKm,
    bool hasApiary = false,
    bool mine = false,
    int limit = 20,
    int offset = 0,
  }) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/listings',
        queryParameters: {
          if (category != null) 'category': category,
          if (keyword != null) 'keyword': keyword,
          if (priceMin != null) 'price_min': priceMin,
          if (priceMax != null) 'price_max': priceMax,
          if (postedAfter != null) 'posted_after': postedAfter,
          if (nearLat != null) 'near_lat': nearLat,
          if (nearLng != null) 'near_lng': nearLng,
          if (radiusKm != null) 'radius_km': radiusKm,
          if (hasApiary) 'has_apiary': true,
          if (mine) 'mine': true,
          'limit': limit,
          'offset': offset,
        },
      );
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>? ?? [])
          .map((e) => Listing.fromJson(e as Map<String, dynamic>))
          .toList();
      return ListingSearchResult(
        items: items,
        total: data['total'] as int? ?? items.length,
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Listing> getListing(int id) async {
    try {
      final response = await _api.dio.get('/api/v1/listings/$id');
      return Listing.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Listing> createListing({
    required String title,
    required String description,
    required String category,
    double? price,
    required String quantity,
    required String address,
    required double lat,
    required double lng,
    int? apiaryId,
    required String contactPhone,
    required String contactEmail,
    int? honeyBatchId,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/listings',
        data: {
          'title': title,
          'description': description,
          'category': category,
          'price': price,
          'quantity': quantity,
          'address': address,
          'lat': lat,
          'lng': lng,
          'apiary_id': apiaryId,
          'contact_phone': contactPhone,
          'contact_email': contactEmail,
          'honey_batch_id': honeyBatchId,
        },
      );
      return Listing.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Listing> updateListing({
    required int id,
    required String title,
    required String description,
    required String category,
    double? price,
    required String quantity,
    required String address,
    required double lat,
    required double lng,
    int? apiaryId,
    required String contactPhone,
    required String contactEmail,
    int? honeyBatchId,
  }) async {
    try {
      final response = await _api.dio.patch(
        '/api/v1/listings/$id',
        data: {
          'title': title,
          'description': description,
          'category': category,
          'price': price,
          'quantity': quantity,
          'address': address,
          'lat': lat,
          'lng': lng,
          'apiary_id': apiaryId,
          'contact_phone': contactPhone,
          'contact_email': contactEmail,
          'honey_batch_id': honeyBatchId,
        },
      );
      return Listing.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Listing> hideListing(int id, {required bool hidden}) async {
    try {
      final response = await _api.dio.patch(
        '/api/v1/listings/$id/hide',
        data: {'hidden': hidden},
      );
      return Listing.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteListing(int id) async {
    try {
      await _api.dio.delete('/api/v1/listings/$id');
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<ListingImage> uploadImage(
    int listingId,
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
        '/api/v1/listings/$listingId/images',
        data: formData,
        onSendProgress: onSendProgress,
      );
      return ListingImage.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteImage(int listingId, int imageId) async {
    try {
      await _api.dio.delete('/api/v1/listings/$listingId/images/$imageId');
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
