part of 'inspections_cubit.dart';

sealed class InspectionsState {}

final class InspectionsInitial extends InspectionsState {}

final class InspectionsLoading extends InspectionsState {}

final class InspectionsLoaded extends InspectionsState {
  final List<Inspection> inspections;
  InspectionsLoaded(this.inspections);
}

final class InspectionsError extends InspectionsState {}
