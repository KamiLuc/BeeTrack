import 'package:flutter/material.dart';

import '../../../l10n/app_localizations.dart';

const _hiveTypes = ['dadant', 'langstroth', 'top_bar', 'wielkopolski'];

const hiveTypeLabels = {
  'dadant': 'Dadant',
  'langstroth': 'Langstroth',
  'top_bar': 'Top bar',
  'wielkopolski': 'Wielkopolski',
};

class HiveActiveToggle extends StatelessWidget {
  final bool value;
  final ValueChanged<bool> onChanged;

  const HiveActiveToggle({
    super.key,
    required this.value,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(
          value ? l10n.hiveActive : l10n.hiveInactive,
          style: Theme.of(context).textTheme.bodyMedium,
        ),
        const SizedBox(width: 12),
        Switch(value: value, onChanged: onChanged),
      ],
    );
  }
}

class HiveNameField extends StatelessWidget {
  final TextEditingController controller;

  const HiveNameField({super.key, required this.controller});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(labelText: l10n.hiveName),
      textInputAction: TextInputAction.next,
      validator: (v) =>
          (v == null || v.trim().isEmpty) ? l10n.generalRequired : null,
    );
  }
}

class HiveTypeDropdown extends StatelessWidget {
  final String value;
  final ValueChanged<String?> onChanged;

  const HiveTypeDropdown({
    super.key,
    required this.value,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return DropdownButtonFormField<String>(
      value: value,
      decoration: InputDecoration(labelText: l10n.hiveType),
      items: _hiveTypes
          .map((t) => DropdownMenuItem(
                value: t,
                child: Text(hiveTypeLabels[t]!),
              ))
          .toList(),
      onChanged: onChanged,
    );
  }
}
