part of 'honey_batches_cubit.dart';

sealed class HoneyBatchesState {}

final class HoneyBatchesInitial extends HoneyBatchesState {}

final class HoneyBatchesLoading extends HoneyBatchesState {}

final class HoneyBatchesError extends HoneyBatchesState {}

final class HoneyBatchesLoaded extends HoneyBatchesState {
  final List<HoneyBatchModel> batches;
  final int currentPage;
  final int totalPages;

  HoneyBatchesLoaded(
    this.batches, {
    required this.currentPage,
    required this.totalPages,
  });

  HoneyBatchesLoaded copyWith({List<HoneyBatchModel>? batches}) {
    return HoneyBatchesLoaded(
      batches ?? this.batches,
      currentPage: currentPage,
      totalPages: totalPages,
    );
  }
}
