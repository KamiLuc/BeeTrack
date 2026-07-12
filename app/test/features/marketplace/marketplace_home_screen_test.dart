import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/widgets/app_drawer.dart';
import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/features/marketplace/view/marketplace_home_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockApiClient extends Mock implements ApiClient {}
class _MockAuthRepository extends Mock implements AuthRepository {}

Widget _wrap(
  Widget child, {
  ApiClient? apiClient,
  AuthBloc? authBloc,
}) =>
    RepositoryProvider<ApiClient>.value(
      value: apiClient ?? _MockApiClient(),
      child: BlocProvider<AuthBloc>.value(
        value: authBloc ?? AuthBloc(auth: _MockAuthRepository()),
        child: MaterialApp(
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: AppLocalizations.supportedLocales,
          locale: const Locale('en'),
          home: child,
        ),
      ),
    );

void main() {
  group('MarketplaceHomeScreen', () {
    testWidgets('unauthenticated: shows marketplace and drawer with login option',
        (tester) async {
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        authBloc: authBloc,
      ));
      await tester.pump();

      expect(find.text('Marketplace'), findsOneWidget);
      expect(find.text('Coming soon'), findsOneWidget);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      expect(find.widgetWithText(ListTile, 'Log in'), findsOneWidget);
      expect(find.widgetWithText(ListTile, 'Apiaries'), findsOneWidget);
    });

    testWidgets(
        'unauthenticated: while browsing, Marketplace tile is selected and '
        'locked Apiaries tile is not, and tapping Apiaries calls onLogin',
        (tester) async {
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      var loginTapped = false;
      await tester.pumpWidget(_wrap(
        MarketplaceHomeScreen(onLogin: () => loginTapped = true),
        authBloc: authBloc,
      ));
      await tester.pump();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      final apiariesTile =
          tester.widget<ListTile>(find.widgetWithText(ListTile, 'Apiaries'));
      expect(apiariesTile.selected, isFalse);

      final marketplaceTile = tester
          .widget<ListTile>(find.widgetWithText(ListTile, 'Marketplace'));
      expect(marketplaceTile.selected, isTrue);

      await tester.tap(find.widgetWithText(ListTile, 'Apiaries'));
      await tester.pumpAndSettle();

      expect(loginTapped, isTrue);
    });

    testWidgets('unauthenticated: tapping Log in invokes onLogin callback',
        (tester) async {
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      var loginTapped = false;
      await tester.pumpWidget(_wrap(
        MarketplaceHomeScreen(onLogin: () => loginTapped = true),
        authBloc: authBloc,
      ));
      await tester.pump();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Log in'));
      await tester.pumpAndSettle();

      expect(loginTapped, isTrue);
    });

    testWidgets('authenticated: shows marketplace with apiaries option',
        (tester) async {
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        authBloc: authBloc,
      ));
      await tester.pump();

      expect(find.text('Marketplace'), findsOneWidget);
      expect(find.text('Coming soon'), findsOneWidget);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      expect(find.widgetWithText(ListTile, 'Apiaries'), findsOneWidget);
      expect(find.widgetWithText(ListTile, 'Log out'), findsOneWidget);
    });

    testWidgets('re-selecting marketplace just closes drawer',
        (tester) async {
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());
      var selected = <AppSection>[];

      await tester.pumpWidget(_wrap(
        MarketplaceHomeScreen(onSelectSection: selected.add),
        authBloc: authBloc,
      ));
      await tester.pump();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Marketplace'));
      await tester.pumpAndSettle();

      expect(find.text('Coming soon'), findsOneWidget);
      expect(selected, isEmpty);
    });

    testWidgets(
        'authenticated: selecting apiaries from drawer invokes onSelectSection',
        (tester) async {
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());
      var selected = <AppSection>[];

      await tester.pumpWidget(_wrap(
        MarketplaceHomeScreen(onSelectSection: selected.add),
        authBloc: authBloc,
      ));
      await tester.pump();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Apiaries'));
      await tester.pumpAndSettle();

      expect(selected, [AppSection.apiaries]);
    });

    testWidgets('authenticated: tapping Log out dispatches LogoutRequested',
        (tester) async {
      final repo = _MockAuthRepository();
      when(() => repo.logout()).thenAnswer((_) async {});
      final authBloc = AuthBloc(auth: repo)..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        authBloc: authBloc,
      ));
      await tester.pump();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Log out'));
      await tester.pump();

      expect(authBloc.state, isA<AuthUnauthenticated>());
    });
  });
}
