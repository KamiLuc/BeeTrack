import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';

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
    ).format(harvestedAt);

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
            decoration: InputDecoration(labelText: l10n.harvestFrames),
            keyboardType: TextInputType.number,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            validator: (_) {
              final frames = int.tryParse(framesController.text.trim()) ?? 0;
              final half = int.tryParse(halfFramesController.text.trim()) ?? 0;
              return (frames == 0 && half == 0) ? l10n.harvestFramesRequired : null;
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: halfFramesController,
            decoration: InputDecoration(labelText: l10n.harvestHalfFrames),
            keyboardType: TextInputType.number,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: kilogramsController,
            decoration: InputDecoration(labelText: l10n.harvestKilograms),
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            inputFormatters: [
              FilteringTextInputFormatter.allow(RegExp(r'[0-9.,]')),
            ],
            validator: (v) => (v == null || v.trim().isEmpty)
                ? l10n.harvestKilogramsRequired
                : null,
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: notesController,
            decoration: InputDecoration(labelText: l10n.harvestNote),
            maxLines: 3,
            minLines: 1,
          ),
        ],
      ),
    );
  }
}
