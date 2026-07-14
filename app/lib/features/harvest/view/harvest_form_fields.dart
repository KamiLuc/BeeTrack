import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';

import '../../../core/validation/size_tiers.dart';
import '../../../l10n/app_localizations.dart';

class HarvestFormFields extends StatelessWidget {
  final GlobalKey<FormState> formKey;
  final DateTime harvestedAt;
  final TextEditingController framesController;
  final TextEditingController halfFramesController;
  final TextEditingController kilogramsController;
  final TextEditingController notesController;
  final VoidCallback onDateTap;

  const HarvestFormFields({
    super.key,
    required this.formKey,
    required this.harvestedAt,
    required this.framesController,
    required this.halfFramesController,
    required this.kilogramsController,
    required this.notesController,
    required this.onDateTap,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).add_Hm().format(harvestedAt);

    return Form(
      key: formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: onDateTap,
            child: InputDecorator(
              decoration: InputDecoration(
                labelText: l10n.harvestDate,
                suffixIcon: const Icon(Icons.calendar_today, size: 20),
              ),
              child: Text(dateStr),
            ),
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: framesController,
            decoration: InputDecoration(
              labelText: l10n.harvestFrames,
              counterText: SizeTier.tiny.counterText,
            ),
            keyboardType: TextInputType.number,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            maxLength: SizeTier.tiny.maxLength,
            validator: (v) {
              final frames = int.tryParse(framesController.text.trim()) ?? 0;
              final half = int.tryParse(halfFramesController.text.trim()) ?? 0;
              if (frames == 0 && half == 0) return l10n.harvestFramesRequired;
              return validateSizeTier(
                v,
                SizeTier.tiny,
                l10n.harvestFrames,
                l10n,
              );
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: halfFramesController,
            decoration: InputDecoration(
              labelText: l10n.harvestHalfFrames,
              counterText: SizeTier.tiny.counterText,
            ),
            keyboardType: TextInputType.number,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            maxLength: SizeTier.tiny.maxLength,
            validator: (v) => validateSizeTier(
              v,
              SizeTier.tiny,
              l10n.harvestHalfFrames,
              l10n,
            ),
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: kilogramsController,
            decoration: InputDecoration(
              labelText: l10n.harvestKilograms,
              counterText: SizeTier.small.counterText,
            ),
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            inputFormatters: [
              FilteringTextInputFormatter.allow(RegExp(r'[0-9.,]')),
            ],
            maxLength: SizeTier.small.maxLength,
            validator: (v) {
              if (v == null || v.trim().isEmpty) {
                return l10n.harvestKilogramsRequired;
              }
              final lengthError = validateSizeTier(
                v,
                SizeTier.small,
                l10n.harvestKilograms,
                l10n,
              );
              if (lengthError != null) return lengthError;
              final parsed = double.tryParse(v.trim().replaceAll(',', '.'));
              if (parsed != null && parsed > maxHarvestKilograms) {
                return l10n.generalValueTooLarge(
                  l10n.harvestKilograms,
                  maxHarvestKilograms.toStringAsFixed(0),
                );
              }
              return null;
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: notesController,
            decoration: InputDecoration(labelText: l10n.harvestNote),
            maxLines: 3,
            minLines: 1,
            maxLength: SizeTier.extraLarge.maxLength,
            validator: (v) => validateSizeTier(
              v,
              SizeTier.extraLarge,
              l10n.harvestNote,
              l10n,
            ),
          ),
        ],
      ),
    );
  }
}
