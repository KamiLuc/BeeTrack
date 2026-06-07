part of 'inspections_cubit.dart';

sealed class InspectionsState {}

final class InspectionsInitial extends InspectionsState {}

final class InspectionsLoading extends InspectionsState {}

final class InspectionsLoaded extends InspectionsState {
  final List<Inspection> inspections;
  final int currentPage;
  final int totalPages;

  InspectionsLoaded(this.inspections, {
    required this.currentPage,
    required this.totalPages,
  });
}

final class InspectionsError extends InspectionsState {}
