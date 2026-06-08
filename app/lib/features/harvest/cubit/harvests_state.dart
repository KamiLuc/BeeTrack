part of 'harvests_cubit.dart';

sealed class HarvestsState {}

final class HarvestsInitial extends HarvestsState {}

final class HarvestsLoading extends HarvestsState {}

final class HarvestsLoaded extends HarvestsState {
  final List<Harvest> harvests;
  final int currentPage;
  final int totalPages;

  HarvestsLoaded(
    this.harvests, {
    required this.currentPage,
    required this.totalPages,
  });
}

final class HarvestsError extends HarvestsState {}
