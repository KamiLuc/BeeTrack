part of 'auth_bloc.dart';

sealed class AuthEvent {}

final class AppStarted extends AuthEvent {}

final class LoginSubmitted extends AuthEvent {
  final String email;
  final String password;

  LoginSubmitted({required this.email, required this.password});
}

final class RegisterSubmitted extends AuthEvent {
  final String email;
  final String lang;
  final String name;
  final String password;

  RegisterSubmitted({
    required this.email,
    required this.lang,
    required this.name,
    required this.password,
  });
}

final class LogoutRequested extends AuthEvent {}
