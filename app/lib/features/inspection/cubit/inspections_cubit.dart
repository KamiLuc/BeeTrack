import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/inspection_model.dart';
import '../data/inspection_repository.dart';

part 'inspections_state.dart';

class InspectionsCubit extends Cubit<InspectionsState> {
  final InspectionRepository _repo;
  final int apiaryId;
  final int hiveId;

  InspectionsCubit({
    required InspectionRepository repo,
    required this.apiaryId,
    required this.hiveId,
  })  : _repo = repo,
        super(InspectionsInitial());

  Future<void> load() async {
    emit(InspectionsLoading());
    try {
      final inspections = await _repo.listInspections(apiaryId, hiveId);
      emit(InspectionsLoaded(inspections));
    } catch (_) {
      emit(InspectionsError());
    }
  }

  Future<void> delete(int inspectionId) async {
    emit(InspectionsLoading());
    try {
      await _repo.deleteInspection(
        apiaryId: apiaryId,
        hiveId: hiveId,
        inspectionId: inspectionId,
      );
      final inspections = await _repo.listInspections(apiaryId, hiveId);
      emit(InspectionsLoaded(inspections));
    } catch (_) {
      emit(InspectionsError());
    }
  }
}
