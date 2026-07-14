import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'listing_model.dart';

class FavoritesRepository {
  final ApiClient _api;

  FavoritesRepository({required this._api});

  Future<void> addFavorite(int listingId) async {
    try {
      await _api.dio.post('/api/v1/listings/$listingId/favorite');
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> removeFavorite(int listingId) async {
    try {
      await _api.dio.delete('/api/v1/listings/$listingId/favorite');
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<List<Listing>> listFavorites() async {
    try {
      final response = await _api.dio.get('/api/v1/favorites');
      final data = response.data as Map<String, dynamic>;
      return (data['items'] as List<dynamic>? ?? [])
          .map((e) => Listing.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<bool> checkFavorite(int listingId) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/listings/$listingId/favorite',
      );
      final data = response.data as Map<String, dynamic>;
      return data['is_favorite'] as bool? ?? false;
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
