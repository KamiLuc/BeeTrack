part of 'feedings_cubit.dart';

sealed class FeedingsState {}

final class FeedingsInitial extends FeedingsState {}

final class FeedingsLoading extends FeedingsState {}

final class FeedingsLoaded extends FeedingsState {
  final List<Feeding> feedings;
  final int currentPage;
  final int totalPages;

  FeedingsLoaded(
    this.feedings, {
    required this.currentPage,
    required this.totalPages,
  });
}

final class FeedingsError extends FeedingsState {}
