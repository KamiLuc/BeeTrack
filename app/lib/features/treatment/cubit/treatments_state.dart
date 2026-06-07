part of 'treatments_cubit.dart';

sealed class TreatmentsState {}

final class TreatmentsInitial extends TreatmentsState {}

final class TreatmentsLoading extends TreatmentsState {}

final class TreatmentsLoaded extends TreatmentsState {
  final List<Treatment> treatments;
  final int currentPage;
  final int totalPages;

  TreatmentsLoaded(
    this.treatments, {
    required this.currentPage,
    required this.totalPages,
  });
}

final class TreatmentsError extends TreatmentsState {}
