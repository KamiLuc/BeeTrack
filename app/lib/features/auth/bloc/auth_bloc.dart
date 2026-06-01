import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_exception.dart';
import '../data/auth_repository.dart';

part 'auth_event.dart';
part 'auth_state.dart';

class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final AuthRepository _auth;

  AuthBloc({required this._auth}) : super(AuthInitial()) {
    on<LoginSubmitted>(_onLoginSubmitted);
    on<RegisterSubmitted>(_onRegisterSubmitted);
    on<LogoutRequested>(_onLogoutRequested);
  }

  Future<void> _onLoginSubmitted(
    LoginSubmitted event,
    Emitter<AuthState> emit,
  ) async {
    emit(AuthLoading());
    try {
      await _auth.login(email: event.email, password: event.password);
      emit(AuthAuthenticated());
    } on ApiException catch (e) {
      emit(AuthFailure(e.code));
    }
  }

  Future<void> _onRegisterSubmitted(
    RegisterSubmitted event,
    Emitter<AuthState> emit,
  ) async {
    emit(AuthLoading());
    try {
      await _auth.register(
        email: event.email,
        name: event.name,
        password: event.password,
      );
      await _auth.login(email: event.email, password: event.password);
      emit(AuthAuthenticated());
    } on ApiException catch (e) {
      emit(AuthFailure(e.code));
    }
  }

  Future<void> _onLogoutRequested(
    LogoutRequested event,
    Emitter<AuthState> emit,
  ) async {
    await _auth.logout();
    emit(AuthUnauthenticated());
  }
}
