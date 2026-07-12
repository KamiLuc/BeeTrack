import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../features/auth/bloc/auth_bloc.dart';
import '../../l10n/app_localizations.dart';
import '../theme/app_colors.dart';

enum AppSection { apiaries, marketplace }

/// A navigation-drawer list tile with an explicit, consistent "active
/// section" highlight — used by both drawers so the selected indicator
/// always looks the same regardless of auth state.
class _DrawerNavTile extends StatelessWidget {
  final IconData icon;
  final String label;
  final bool selected;
  final Widget? trailing;
  final VoidCallback onTap;

  const _DrawerNavTile({
    required this.icon,
    required this.label,
    required this.selected,
    this.trailing,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Icon(icon),
      title: Text(label),
      trailing: trailing,
      selected: selected,
      selectedColor: AppColors.primaryDark,
      selectedTileColor: AppColors.primary.withValues(alpha: 0.15),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      onTap: onTap,
    );
  }
}

/// A navigation drawer for switching between top-level sections.
/// Shown on authenticated screens (Apiaries + Marketplace).
/// [current] marks the active section; [onSelect] is called with the chosen
/// section. Section switching is state-driven by the host shell, not routed.
class AuthenticatedAppDrawer extends StatelessWidget {
  final AppSection current;
  final ValueChanged<AppSection> onSelect;

  const AuthenticatedAppDrawer({
    super.key,
    required this.current,
    required this.onSelect,
  });

  void _select(BuildContext context, AppSection target) {
    Navigator.of(context).pop();
    if (target != current) onSelect(target);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Drawer(
      child: SafeArea(
        child: Column(
          children: [
            DrawerHeader(
              child: Align(
                alignment: Alignment.bottomLeft,
                child: Text(
                  l10n.appName,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
              ),
            ),
            _DrawerNavTile(
              icon: Icons.hive_outlined,
              label: l10n.apiaryTitle,
              selected: current == AppSection.apiaries,
              onTap: () => _select(context, AppSection.apiaries),
            ),
            _DrawerNavTile(
              icon: Icons.storefront_outlined,
              label: l10n.marketplaceTitle,
              selected: current == AppSection.marketplace,
              onTap: () => _select(context, AppSection.marketplace),
            ),
            const Spacer(),
            ListTile(
              leading: const Icon(Icons.logout),
              title: Text(l10n.authLogout),
              onTap: () {
                Navigator.of(context).pop();
                context.read<AuthBloc>().add(LogoutRequested());
              },
            ),
          ],
        ),
      ),
    );
  }
}

/// A navigation drawer for unauthenticated users, shown on both the public
/// marketplace and the login screen. Marketplace is browsable; Apiaries is
/// shown but gated — tapping it (or the login entry) calls [onLogin].
/// [isLogin] marks whether the login view is currently shown.
class UnauthenticatedAppDrawer extends StatelessWidget {
  final bool isLogin;
  final VoidCallback onMarketplace;
  final VoidCallback onLogin;

  const UnauthenticatedAppDrawer({
    super.key,
    this.isLogin = false,
    required this.onMarketplace,
    required this.onLogin,
  });

  void _go(BuildContext context, VoidCallback action, {required bool isCurrent}) {
    Navigator.of(context).pop();
    if (!isCurrent) action();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Drawer(
      child: SafeArea(
        child: Column(
          children: [
            DrawerHeader(
              child: Align(
                alignment: Alignment.bottomLeft,
                child: Text(
                  l10n.appName,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
              ),
            ),
            _DrawerNavTile(
              icon: Icons.hive_outlined,
              label: l10n.apiaryTitle,
              trailing: const Icon(Icons.lock_outline, size: 18),
              selected: false,
              onTap: () => _go(context, onLogin, isCurrent: false),
            ),
            _DrawerNavTile(
              icon: Icons.storefront_outlined,
              label: l10n.marketplaceTitle,
              selected: !isLogin,
              onTap: () => _go(context, onMarketplace, isCurrent: !isLogin),
            ),
            const Spacer(),
            _DrawerNavTile(
              icon: Icons.login,
              label: l10n.authLogin,
              selected: isLogin,
              onTap: () => _go(context, onLogin, isCurrent: isLogin),
            ),
          ],
        ),
      ),
    );
  }
}
