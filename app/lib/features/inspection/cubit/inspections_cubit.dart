import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/inspection_model.dart';
import '../data/inspection_repository.dart';

part 'inspections_state.dart';

const _pageSize = 10;

class InspectionsCubit extends Cubit<InspectionsState> {
  final InspectionRepository _repo;
  final int apiaryId;
  final int hiveId;

  InspectionsCubit({
    required InspectionRepository repo,
    required this.apiaryId,
    required this.hiveId,
  })  : _repo = repo,
        super(InspectionsInitial());

  Future<void> load() => _goToPage(1);

  Future<void> goToPage(int page) => _goToPage(page);

  Future<void> _goToPage(int page) async {
    emit(InspectionsLoading());
    try {
      final result = await _repo.listInspections(
        apiaryId,
        hiveId,
        limit: _pageSize,
        offset: (page - 1) * _pageSize,
      );
      final totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
      emit(InspectionsLoaded(
        result.items,
        currentPage: page,
        totalPages: totalPages,
      ));
    } catch (_) {
      emit(InspectionsError());
    }
  }

  Future<void> delete(int inspectionId) async {
    final currentPage =
        state is InspectionsLoaded ? (state as InspectionsLoaded).currentPage : 1;
    emit(InspectionsLoading());
    try {
      await _repo.deleteInspection(
        apiaryId: apiaryId,
        hiveId: hiveId,
        inspectionId: inspectionId,
      );
      final result = await _repo.listInspections(
        apiaryId,
        hiveId,
        limit: _pageSize,
        offset: (currentPage - 1) * _pageSize,
      );
      final totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
      final page = result.items.isEmpty && currentPage > 1
          ? currentPage - 1
          : currentPage;
      if (page != currentPage) {
        await _goToPage(page);
      } else {
        emit(InspectionsLoaded(
          result.items,
          currentPage: page,
          totalPages: totalPages,
        ));
      }
    } catch (_) {
      emit(InspectionsError());
    }
  }
}
