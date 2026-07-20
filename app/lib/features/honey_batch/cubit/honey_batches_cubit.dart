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
}
