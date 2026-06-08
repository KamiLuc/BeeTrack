import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/harvest_model.dart';
import '../data/harvest_repository.dart';

part 'harvests_state.dart';

const _pageSize = 10;

class HarvestsCubit extends Cubit<HarvestsState> {
  final HarvestRepository _repo;
  final int apiaryId;
  final int hiveId;

  HarvestsCubit({
    required HarvestRepository repo,
    required this.apiaryId,
    required this.hiveId,
  })  : _repo = repo,
        super(HarvestsInitial());

  Future<void> load() => _goToPage(1);

  Future<void> goToPage(int page) => _goToPage(page);

  Future<void> _goToPage(int page) async {
    emit(HarvestsLoading());
    try {
      final result = await _repo.listHarvests(
        apiaryId,
        hiveId,
        limit: _pageSize,
        offset: (page - 1) * _pageSize,
      );
      final totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
      emit(HarvestsLoaded(
        result.items,
        currentPage: page,
        totalPages: totalPages,
      ));
    } catch (_) {
      emit(HarvestsError());
    }
  }

  Future<void> delete(int harvestId) async {
    final currentPage =
        state is HarvestsLoaded ? (state as HarvestsLoaded).currentPage : 1;
    emit(HarvestsLoading());
    try {
      await _repo.deleteHarvest(
        apiaryId: apiaryId,
        hiveId: hiveId,
        harvestId: harvestId,
      );
      final result = await _repo.listHarvests(
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
        emit(HarvestsLoaded(
          result.items,
          currentPage: page,
          totalPages: totalPages,
        ));
      }
    } catch (_) {
      emit(HarvestsError());
    }
  }
}
