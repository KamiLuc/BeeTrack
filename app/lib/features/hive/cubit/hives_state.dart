part of 'hives_cubit.dart';

sealed class HivesState {}

final class HivesInitial extends HivesState {}

final class HivesLoading extends HivesState {}

final class HivesLoaded extends HivesState {
  final List<Hive> hives;
  HivesLoaded(this.hives);
}

final class HivesError extends HivesState {}
