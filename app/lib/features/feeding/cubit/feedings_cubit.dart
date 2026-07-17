import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/feeding_model.dart';
import '../data/feeding_repository.dart';

part 'feedings_state.dart';

const _pageSize = 10;

class FeedingsCubit extends Cubit<FeedingsState> {
  final FeedingRepository _repo;
  final int apiaryId;
  final int hiveId;

  FeedingsCubit({
    required FeedingRepository repo,
    required this.apiaryId,
    required this.hiveId,
  })  : _repo = repo,
        super(FeedingsInitial());

  Future<void> load() => _goToPage(1);

  Future<void> goToPage(int page) => _goToPage(page);

  Future<void> _goToPage(int page) async {
    emit(FeedingsLoading());
    try {
      final result = await _repo.listFeedings(
        apiaryId,
        hiveId,
        limit: _pageSize,
        offset: (page - 1) * _pageSize,
      );
      final totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
      emit(FeedingsLoaded(
        result.items,
        currentPage: page,
        totalPages: totalPages,
      ));
    } catch (_) {
      emit(FeedingsError());
    }
  }

  Future<void> delete(int feedingId) async {
    final currentPage =
        state is FeedingsLoaded ? (state as FeedingsLoaded).currentPage : 1;
    emit(FeedingsLoading());
    try {
      await _repo.deleteFeeding(
        apiaryId: apiaryId,
        hiveId: hiveId,
        feedingId: feedingId,
      );
      final result = await _repo.listFeedings(
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
        emit(FeedingsLoaded(
          result.items,
          currentPage: page,
          totalPages: totalPages,
        ));
      }
    } catch (_) {
      emit(FeedingsError());
    }
  }
}
