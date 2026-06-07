import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/treatment_model.dart';
import '../data/treatment_repository.dart';

part 'treatments_state.dart';

const _pageSize = 10;

class TreatmentsCubit extends Cubit<TreatmentsState> {
  final TreatmentRepository _repo;
  final int apiaryId;
  final int hiveId;

  TreatmentsCubit({
    required TreatmentRepository repo,
    required this.apiaryId,
    required this.hiveId,
  })  : _repo = repo,
        super(TreatmentsInitial());

  Future<void> load() => _goToPage(1);

  Future<void> goToPage(int page) => _goToPage(page);

  Future<void> _goToPage(int page) async {
    emit(TreatmentsLoading());
    try {
      final result = await _repo.listTreatments(
        apiaryId,
        hiveId,
        limit: _pageSize,
        offset: (page - 1) * _pageSize,
      );
      final totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
      emit(TreatmentsLoaded(
        result.items,
        currentPage: page,
        totalPages: totalPages,
      ));
    } catch (_) {
      emit(TreatmentsError());
    }
  }

  Future<void> delete(int treatmentId) async {
    final currentPage =
        state is TreatmentsLoaded ? (state as TreatmentsLoaded).currentPage : 1;
    emit(TreatmentsLoading());
    try {
      await _repo.deleteTreatment(
        apiaryId: apiaryId,
        hiveId: hiveId,
        treatmentId: treatmentId,
      );
      final result = await _repo.listTreatments(
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
        emit(TreatmentsLoaded(
          result.items,
          currentPage: page,
          totalPages: totalPages,
        ));
      }
    } catch (_) {
      emit(TreatmentsError());
    }
  }
}
