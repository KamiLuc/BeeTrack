import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'inspection_model.dart';

class InspectionRepository {
  final ApiClient _api;

  InspectionRepository({required ApiClient api}) : _api = api;

  Future<({List<Inspection> items, int total})> listInspections(
    int apiaryId,
    int hiveId, {
    int limit = 20,
    int offset = 0,
  }) async {
    try {
      final response = await _api.dio.get(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/inspections',
        queryParameters: {'limit': limit, 'offset': offset},
      );
      final body = response.data as Map<String, dynamic>;
      final items = (body['items'] as List<dynamic>)
          .map((e) => Inspection.fromJson(e as Map<String, dynamic>))
          .toList();
      final total = body['total'] as int;
      return (items: items, total: total);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Inspection> createInspection({
    required int apiaryId,
    required int hiveId,
    required DateTime inspectedAt,
    required String queenSeen,
    required String broodPattern,
    required String aggressiveness,
    required bool queenAdded,
    required String notes,
    int? framesBrood,
    int? framesHoney,
    int? framesPollen,
    int? framesAddedDrawn,
    int? framesAddedFoundation,
    int? framesAddedHoney,
    int? queenCellsCount,
  }) async {
    try {
      final response = await _api.dio.post(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/inspections',
        data: {
          'aggressiveness': aggressiveness,
          'brood_pattern': broodPattern,
          'frames_added_drawn': framesAddedDrawn,
          'frames_added_foundation': framesAddedFoundation,
          'frames_added_honey': framesAddedHoney,
          'frames_brood': framesBrood,
          'frames_honey': framesHoney,
          'frames_pollen': framesPollen,
          'inspected_at': inspectedAt.toUtc().toIso8601String(),
          'notes': notes,
          'queen_added': queenAdded,
          'queen_cells_count': queenCellsCount,
          'queen_status': queenSeen,
        },
      );
      return Inspection.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<Inspection> updateInspection({
    required int apiaryId,
    required int hiveId,
    required int inspectionId,
    required DateTime inspectedAt,
    required String queenSeen,
    required String broodPattern,
    required String aggressiveness,
    required bool queenAdded,
    required String notes,
    int? framesBrood,
    int? framesHoney,
    int? framesPollen,
    int? framesAddedDrawn,
    int? framesAddedFoundation,
    int? framesAddedHoney,
    int? queenCellsCount,
  }) async {
    try {
      final response = await _api.dio.patch(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/inspections/$inspectionId',
        data: {
          'aggressiveness': aggressiveness,
          'brood_pattern': broodPattern,
          'frames_added_drawn': framesAddedDrawn,
          'frames_added_foundation': framesAddedFoundation,
          'frames_added_honey': framesAddedHoney,
          'frames_brood': framesBrood,
          'frames_honey': framesHoney,
          'frames_pollen': framesPollen,
          'inspected_at': inspectedAt.toUtc().toIso8601String(),
          'notes': notes,
          'queen_added': queenAdded,
          'queen_cells_count': queenCellsCount,
          'queen_status': queenSeen,
        },
      );
      return Inspection.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteInspection({
    required int apiaryId,
    required int hiveId,
    required int inspectionId,
  }) async {
    try {
      await _api.dio.delete(
        '/api/v1/apiaries/$apiaryId/hives/$hiveId/inspections/$inspectionId',
      );
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }
}
