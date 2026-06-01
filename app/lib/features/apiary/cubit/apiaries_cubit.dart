import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/apiary_model.dart';
import '../data/apiary_repository.dart';

part 'apiaries_state.dart';

class ApiariesCubit extends Cubit<ApiariesState> {
  final ApiaryRepository _repo;

  ApiariesCubit({required this._repo}) : super(ApiariesInitial());

  Future<void> load() async {
    emit(ApiariesLoading());
    try {
      final apiaries = await _repo.listApiaries();
      emit(ApiariesLoaded(apiaries));
    } catch (_) {
      emit(ApiariesError());
    }
  }
}
