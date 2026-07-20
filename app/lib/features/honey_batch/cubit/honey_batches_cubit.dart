import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/honey_batch_model.dart';
import '../data/honey_batch_repository.dart';
import '../data/processing_method.dart';

part 'honey_batches_state.dart';

const _pageSize = 20;

class HoneyBatchesCubit extends Cubit<HoneyBatchesState> {
  final HoneyBatchRepository _repo;

  HoneyBatchesCubit({required HoneyBatchRepository repo})
      : _repo = repo,
        super(HoneyBatchesInitial());

  Future<void> load() => _goToPage(1);
  Future<void> goToPage(int page) => _goToPage(page);

  Future<void> _goToPage(int page) async {
    emit(HoneyBatchesLoading());
    try {
      final result = await _repo.listBatches(
        limit: _pageSize,
        offset: (page - 1) * _pageSize,
      );
      final totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
      emit(
        HoneyBatchesLoaded(
          result.items,
          currentPage: page,
          totalPages: totalPages,
        ),
      );
    } catch (_) {
      emit(HoneyBatchesError());
    }
  }

  /// Creates a batch and reloads the first page. `certification` on the
  /// returned batch is `queued` when [requestCertification] is true, or
  /// `null` when it's left false. Rethrows on failure so the caller can
  /// surface the specific error.
  Future<HoneyBatchModel> create({
    required DateTime gatheringDate,
    required int amountGrams,
    required ProcessingMethod processingMethod,
    required String honeyType,
    List<int>? pdfBytes,
    String? pdfFilename,
    bool requestCertification = false,
  }) async {
    final batch = await _repo.createBatch(
      gatheringDate: gatheringDate,
      amountGrams: amountGrams,
      processingMethod: processingMethod,
      honeyType: honeyType,
      pdfBytes: pdfBytes,
      pdfFilename: pdfFilename,
      requestCertification: requestCertification,
    );
    await _goToPage(1);
    return batch;
  }

  /// Certifies a batch for the first time or retries a failed/reverted one.
  /// Replaces the matching item in the current list on success; rethrows
  /// on failure without touching state so the caller can show the error.
  Future<void> requestCertification(int id) async {
    final batch = await _repo.requestCertification(id);
    final current = state;
    if (current is! HoneyBatchesLoaded) return;
    final items = current.batches
        .map((b) => b.id == id ? batch : b)
        .toList();
    emit(current.copyWith(batches: items));
  }

  /// Replaces a single batch already in the current list with an updated
  /// copy — used after a screen that made its own repository call (e.g. the
  /// edit form, which runs outside this cubit's provider scope) returns the
  /// fresh batch. No-ops if state isn't Loaded.
  void replaceInList(HoneyBatchModel batch) {
    final current = state;
    if (current is! HoneyBatchesLoaded) return;
    final items = current.batches
        .map((b) => b.id == batch.id ? batch : b)
        .toList();
    emit(current.copyWith(batches: items));
  }

  /// Deletes a batch and reloads the current page, stepping back a page if
  /// that page is now empty. Rethrows on failure without touching state.
  Future<void> delete(int id) async {
    final current = state;
    if (current is! HoneyBatchesLoaded) return;
    final currentPage = current.currentPage;
    await _repo.deleteBatch(id);
    final result = await _repo.listBatches(
      limit: _pageSize,
      offset: (currentPage - 1) * _pageSize,
    );
    final page = result.items.isEmpty && currentPage > 1
        ? currentPage - 1
        : currentPage;
    if (page != currentPage) {
      await _goToPage(page);
    } else {
      final totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
      emit(
        HoneyBatchesLoaded(result.items, currentPage: page, totalPages: totalPages),
      );
    }
  }
}
