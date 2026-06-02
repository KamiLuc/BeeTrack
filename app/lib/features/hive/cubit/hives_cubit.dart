import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/hive_model.dart';
import '../data/hive_repository.dart';

part 'hives_state.dart';

class HivesCubit extends Cubit<HivesState> {
  final HiveRepository _repo;
  final int apiaryId;

  HivesCubit({required this._repo, required this.apiaryId})
      : super(HivesInitial());

  Future<void> load() async {
    emit(HivesLoading());
    try {
      final hives = await _repo.listHives(apiaryId);
      emit(HivesLoaded(hives));
    } catch (_) {
      emit(HivesError());
    }
  }

  Future<void> delete(int hiveId) async {
    emit(HivesLoading());
    try {
      await _repo.deleteHive(apiaryId: apiaryId, hiveId: hiveId);
      final hives = await _repo.listHives(apiaryId);
      emit(HivesLoaded(hives));
    } catch (_) {
      emit(HivesError());
    }
  }
}
