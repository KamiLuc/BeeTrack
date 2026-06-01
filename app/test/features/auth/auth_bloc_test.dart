import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/core/api/api_exception.dart';
import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';

class MockAuthRepository extends Mock implements AuthRepository {}

void main() {
  late MockAuthRepository repo;
  late AuthBloc bloc;

  setUp(() {
    repo = MockAuthRepository();
    bloc = AuthBloc(auth: repo);
  });

  tearDown(() => bloc.close());

  group('AppStarted', () {
    blocTest<AuthBloc, AuthState>(
      'emits [AuthAuthenticated] when already logged in',
      build: () {
        when(() => repo.isLoggedIn).thenReturn(true);
        return bloc;
      },
      act: (b) => b.add(AppStarted()),
      expect: () => [isA<AuthAuthenticated>()],
    );

    blocTest<AuthBloc, AuthState>(
      'emits [AuthUnauthenticated] when not logged in',
      build: () {
        when(() => repo.isLoggedIn).thenReturn(false);
        return bloc;
      },
      act: (b) => b.add(AppStarted()),
      expect: () => [isA<AuthUnauthenticated>()],
    );
  });

  group('LoginSubmitted', () {
    blocTest<AuthBloc, AuthState>(
      'emits [AuthLoading, AuthAuthenticated] on success',
      build: () {
        when(
          () => repo.login(email: 'a@b.com', password: 'password123'),
        ).thenAnswer((_) async {});
        return bloc;
      },
      act: (b) => b.add(
        LoginSubmitted(email: 'a@b.com', password: 'password123'),
      ),
      expect: () => [isA<AuthLoading>(), isA<AuthAuthenticated>()],
    );

    blocTest<AuthBloc, AuthState>(
      'emits [AuthLoading, AuthFailure] on invalid credentials',
      build: () {
        when(
          () => repo.login(email: any(named: 'email'), password: any(named: 'password')),
        ).thenThrow(
          const ApiException(code: 'INVALID_CREDENTIALS', message: 'wrong'),
        );
        return bloc;
      },
      act: (b) => b.add(
        LoginSubmitted(email: 'a@b.com', password: 'wrongpass'),
      ),
      expect: () => [
        isA<AuthLoading>(),
        isA<AuthFailure>().having((s) => s.code, 'code', 'INVALID_CREDENTIALS'),
      ],
    );
  });

  group('RegisterSubmitted', () {
    blocTest<AuthBloc, AuthState>(
      'emits [AuthLoading, AuthAuthenticated] on success',
      build: () {
        when(
          () => repo.register(
            email: any(named: 'email'),
            name: any(named: 'name'),
            password: any(named: 'password'),
          ),
        ).thenAnswer((_) async {});
        when(
          () => repo.login(
            email: any(named: 'email'),
            password: any(named: 'password'),
          ),
        ).thenAnswer((_) async {});
        return bloc;
      },
      act: (b) => b.add(
        RegisterSubmitted(
          email: 'a@b.com',
          name: 'Alice',
          password: 'password123',
        ),
      ),
      expect: () => [isA<AuthLoading>(), isA<AuthAuthenticated>()],
    );

    blocTest<AuthBloc, AuthState>(
      'emits [AuthLoading, AuthFailure] when email is taken',
      build: () {
        when(
          () => repo.register(
            email: any(named: 'email'),
            name: any(named: 'name'),
            password: any(named: 'password'),
          ),
        ).thenThrow(
          const ApiException(code: 'EMAIL_TAKEN', message: 'taken'),
        );
        return bloc;
      },
      act: (b) => b.add(
        RegisterSubmitted(
          email: 'taken@b.com',
          name: 'Alice',
          password: 'password123',
        ),
      ),
      expect: () => [
        isA<AuthLoading>(),
        isA<AuthFailure>().having((s) => s.code, 'code', 'EMAIL_TAKEN'),
      ],
    );
  });

  group('LogoutRequested', () {
    blocTest<AuthBloc, AuthState>(
      'emits [AuthUnauthenticated]',
      build: () {
        when(() => repo.logout()).thenAnswer((_) async {});
        return bloc;
      },
      act: (b) => b.add(LogoutRequested()),
      expect: () => [isA<AuthUnauthenticated>()],
    );
  });
}
