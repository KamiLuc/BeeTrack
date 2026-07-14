import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/core/widgets/app_drawer.dart';
import 'package:app/features/apiary/view/apiaries_screen.dart';
import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/features/auth/view/login_screen.dart';
import 'package:app/features/marketplace/view/marketplace_home_screen.dart';
import 'package:app/l10n/app_localizations.dart';
import 'package:app/main.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

/// Fails every request synchronously so tests never leave a pending network
/// timer behind when a screen (e.g. ApiariesScreen) eagerly loads data.
class _FailingHttpClientAdapter implements HttpClientAdapter {
  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) {
    throw DioException(requestOptions: options, message: 'no network in tests');
  }
}

/// The [TokenStorage] backing the current test's [ApiClient], also handed to
/// [_pumpAuthWrapper] via a [RepositoryProvider] so `context.read<TokenStorage>()`
/// resolves just like it does in the real widget tree (see main.dart).
late TokenStorage _tokenStorage;

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  _tokenStorage = TokenStorage(prefs);
  final apiClient = ApiClient(storage: _tokenStorage, baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = _FailingHttpClientAdapter();
  return apiClient;
}

Future<AuthBloc> _pumpAuthWrapper(
  WidgetTester tester, {
  required ApiClient apiClient,
  required AuthBloc authBloc,
}) async {
  await tester.pumpWidget(
    RepositoryProvider<ApiClient>.value(
      value: apiClient,
      child: RepositoryProvider<TokenStorage>.value(
        value: _tokenStorage,
        child: BlocProvider<AuthBloc>.value(
          value: authBloc,
          child: MaterialApp(
            localizationsDelegates: AppLocalizations.localizationsDelegates,
            supportedLocales: AppLocalizations.supportedLocales,
            locale: const Locale('en'),
            home: const AuthWrapper(),
          ),
        ),
      ),
    ),
  );
  await tester.pump();
  return authBloc;
}

void main() {
  group('AuthWrapper', () {
    testWidgets('cold start: unauthenticated user lands on the login gate',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthUnauthenticated());

      await _pumpAuthWrapper(tester, apiClient: apiClient, authBloc: authBloc);

      expect(find.byType(LoginScreen), findsOneWidget);
      expect(find.byType(MarketplaceHomeScreen), findsNothing);
    });

    testWidgets(
        'tapping Marketplace in login drawer returns to browsing',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthUnauthenticated());

      await _pumpAuthWrapper(tester, apiClient: apiClient, authBloc: authBloc);
      expect(find.byType(LoginScreen), findsOneWidget);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ListTile, 'Marketplace'));
      await tester.pumpAndSettle();

      expect(find.byType(MarketplaceHomeScreen), findsOneWidget);
      expect(find.byType(LoginScreen), findsNothing);
    });

    testWidgets(
        'tapping locked Apiaries from marketplace goes to login',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthUnauthenticated());

      await _pumpAuthWrapper(tester, apiClient: apiClient, authBloc: authBloc);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ListTile, 'Marketplace'));
      await tester.pumpAndSettle();
      expect(find.byType(MarketplaceHomeScreen), findsOneWidget);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ListTile, 'Apiaries'));
      await tester.pumpAndSettle();

      expect(find.byType(LoginScreen), findsOneWidget);
      expect(find.byType(MarketplaceHomeScreen), findsNothing);
    });

    testWidgets(
        'successful login lands on Apiaries with section=apiaries',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthUnauthenticated());

      await _pumpAuthWrapper(tester, apiClient: apiClient, authBloc: authBloc);
      expect(find.byType(LoginScreen), findsOneWidget);

      authBloc.emit(AuthAuthenticated());
      await tester.pumpAndSettle();

      expect(find.byType(ApiariesScreen), findsOneWidget);
      expect(find.widgetWithText(AppBar, 'Apiaries'), findsOneWidget);
    });

    testWidgets(
        'logout from marketplace returns to the login gate, not marketplace',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthUnauthenticated());

      await _pumpAuthWrapper(tester, apiClient: apiClient, authBloc: authBloc);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ListTile, 'Marketplace'));
      await tester.pumpAndSettle();
      expect(find.byType(MarketplaceHomeScreen), findsOneWidget);

      authBloc.emit(AuthUnauthenticated());
      await tester.pumpAndSettle();

      expect(find.byType(LoginScreen), findsOneWidget);
      expect(find.byType(MarketplaceHomeScreen), findsNothing);
    });

    testWidgets(
        'logout from apiaries returns to the login gate, not marketplace',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await _pumpAuthWrapper(tester, apiClient: apiClient, authBloc: authBloc);
      await tester.pumpAndSettle();
      expect(find.byType(ApiariesScreen), findsOneWidget);

      authBloc.emit(AuthUnauthenticated());
      await tester.pumpAndSettle();

      expect(find.byType(LoginScreen), findsOneWidget);
      expect(find.byType(MarketplaceHomeScreen), findsNothing);
    });

    testWidgets(
        'AuthWrapper state persists (never rebuilt via navigation) across '
        'login, section switch, and logout', (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthUnauthenticated());

      await _pumpAuthWrapper(tester, apiClient: apiClient, authBloc: authBloc);
      final initialState = tester.state(find.byType(AuthWrapper));

      authBloc.emit(AuthAuthenticated());
      await tester.pumpAndSettle();
      expect(tester.state(find.byType(AuthWrapper)), same(initialState));

      final apiariesScreen =
          tester.widget<ApiariesScreen>(find.byType(ApiariesScreen));
      apiariesScreen.onSelectSection(AppSection.marketplace);
      await tester.pumpAndSettle();
      expect(tester.state(find.byType(AuthWrapper)), same(initialState));
      expect(find.byType(MarketplaceHomeScreen), findsOneWidget);

      authBloc.emit(AuthUnauthenticated());
      await tester.pumpAndSettle();
      expect(tester.state(find.byType(AuthWrapper)), same(initialState));
      expect(find.byType(LoginScreen), findsOneWidget);
    });
  });
}
