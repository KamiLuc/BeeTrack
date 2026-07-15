import 'package:flutter/material.dart';

import '../../l10n/app_localizations.dart';
import '../validation/size_tiers.dart';

class LocationPickerSection extends StatelessWidget {
  final TextEditingController latController;
  final TextEditingController lngController;
  final String latLabel;
  final String lngLabel;
  final bool locating;
  final VoidCallback onGps;
  final VoidCallback onMap;
  final String? Function(String?)? latValidator;
  final String? Function(String?)? lngValidator;

  const LocationPickerSection({
    super.key,
    required this.latController,
    required this.lngController,
    required this.latLabel,
    required this.lngLabel,
    required this.locating,
    required this.onGps,
    required this.onMap,
    this.latValidator,
    this.lngValidator,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextFormField(
          controller: latController,
          enabled: false,
          decoration: InputDecoration(
            labelText: latLabel,
            counterText: SizeTier.small.counterText,
          ),
          maxLength: SizeTier.small.maxLength,
          validator: latValidator,
        ),
        const SizedBox(height: 12),
        TextFormField(
          controller: lngController,
          enabled: false,
          decoration: InputDecoration(
            labelText: lngLabel,
            counterText: SizeTier.small.counterText,
          ),
          maxLength: SizeTier.small.maxLength,
          validator: lngValidator,
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: locating ? null : onGps,
                icon: locating
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.my_location, size: 18),
                label: Text(l10n.locationPickerGpsButton),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: OutlinedButton.icon(
                onPressed: onMap,
                icon: const Icon(Icons.map, size: 18),
                label: Text(l10n.locationPickerMapButton),
              ),
            ),
          ],
        ),
      ],
    );
  }
}
