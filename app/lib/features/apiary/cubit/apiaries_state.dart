part of 'apiaries_cubit.dart';

sealed class ApiariesState {}

final class ApiariesInitial extends ApiariesState {}

final class ApiariesLoading extends ApiariesState {}

final class ApiariesLoaded extends ApiariesState {
  final List<Apiary> apiaries;
  ApiariesLoaded(this.apiaries);
}

final class ApiariesError extends ApiariesState {}
