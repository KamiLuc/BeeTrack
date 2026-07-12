import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/widgets/app_drawer.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../features/auth/bloc/auth_bloc.dart';
import '../../../l10n/app_localizations.dart';

class MarketplaceHomeScreen extends StatelessWidget {
  /// Called when an authenticated user picks a section from the drawer.
  final ValueChanged<AppSection>? onSelectSection;

  /// Called when an unauthenticated user taps "Log in" in the drawer.
  final VoidCallback? onLogin;

  const MarketplaceHomeScreen({
    super.key,
    this.onSelectSection,
    this.onLogin,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return BlocBuilder<AuthBloc, AuthState>(
      builder: (context, authState) {
        final isAuthenticated = authState is AuthAuthenticated;
        final drawer = isAuthenticated
            ? AuthenticatedAppDrawer(
                current: AppSection.marketplace,
                onSelect: onSelectSection ?? (_) {},
              )
            : UnauthenticatedAppDrawer(
                onMarketplace: () {},
                onLogin: onLogin ?? () {},
              );

        return Scaffold(
          appBar: AppBar(
            title: Text(l10n.marketplaceTitle),
            actions: isAuthenticated ? const [ProfileIconButton()] : null,
          ),
          drawer: drawer,
          body: Center(child: Text(l10n.marketplaceComingSoon)),
        );
      },
    );
  }
}
