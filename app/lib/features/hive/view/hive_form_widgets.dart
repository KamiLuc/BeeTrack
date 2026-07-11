import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../l10n/app_localizations.dart';
import '../data/hive_model.dart';

const _hiveTypes = ['dadant', 'langstroth', 'top_bar', 'wielkopolski'];

const hiveTypeLabels = {
  'dadant': 'Dadant',
  'langstroth': 'Langstroth',
  'top_bar': 'Top bar',
  'wielkopolski': 'Wielkopolski',
};

class _LabeledSwitch extends StatelessWidget {
  final String label;
  final bool value;
  final ValueChanged<bool> onChanged;

  const _LabeledSwitch({
    required this.label,
    required this.value,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: Theme.of(context).textTheme.bodyMedium),
        Switch(value: value, onChanged: onChanged),
      ],
    );
  }
}

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
    return _LabeledSwitch(
      label: value ? l10n.hiveActive : l10n.hiveInactive,
      value: value,
      onChanged: onChanged,
    );
  }
}

class HiveNameField extends StatelessWidget {
  final TextEditingController controller;
  final Set<String> existingNames;

  const HiveNameField({
    super.key,
    required this.controller,
    this.existingNames = const {},
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(labelText: l10n.hiveName),
      textInputAction: TextInputAction.next,
      validator: (v) {
        final trimmed = v?.trim() ?? '';
        if (trimmed.isEmpty) return l10n.generalRequired;
        if (existingNames.contains(trimmed.toLowerCase())) {
          return l10n.hiveDuplicateName;
        }
        return null;
      },
    );
  }
}

class HiveQueenlessToggle extends StatelessWidget {
  final bool value;
  final ValueChanged<bool> onChanged;

  const HiveQueenlessToggle({
    super.key,
    required this.value,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return _LabeledSwitch(
      label: l10n.hiveQueenless,
      value: value,
      onChanged: onChanged,
    );
  }
}

class HiveReadyForHarvestToggle extends StatelessWidget {
  final bool value;
  final ValueChanged<bool> onChanged;

  const HiveReadyForHarvestToggle({
    super.key,
    required this.value,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return _LabeledSwitch(
      label: l10n.hiveReadyForHarvest,
      value: value,
      onChanged: onChanged,
    );
  }
}

String hiveDiseaseLabel(AppLocalizations l10n, String v) => switch (v) {
  'varroa' => l10n.inspectionDiseaseVarroa,
  'nosema' => l10n.inspectionDiseaseNosema,
  'dwv' => l10n.inspectionDiseaseDwv,
  'american_foulbrood' => l10n.inspectionDiseaseAmericanFoulbrood,
  'chalkbrood' => l10n.inspectionDiseaseChalkbrood,
  'european_foulbrood' => l10n.inspectionDiseaseEuropeanFoulbrood,
  'laying_workers' => l10n.inspectionDiseaseLayingWorkers,
  _ => v,
};

class HiveDiseasesSection extends StatelessWidget {
  final String label;
  final Set<String> selected;
  final void Function(String disease, bool selected) onToggle;

  const HiveDiseasesSection({
    super.key,
    required this.label,
    required this.selected,
    required this.onToggle,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Align(
          alignment: Alignment.centerLeft,
          child: Text(label, style: Theme.of(context).textTheme.bodyMedium),
        ),
        const SizedBox(height: 8),
        Wrap(
          alignment: WrapAlignment.center,
          spacing: 8,
          runSpacing: 4,
          children: diseaseValues.map((disease) {
            final isSelected = selected.contains(disease);
            return FilterChip(
              label: Text(hiveDiseaseLabel(l10n, disease)),
              selected: isSelected,
              selectedColor: colorScheme.errorContainer,
              checkmarkColor: colorScheme.onErrorContainer,
              onSelected: (v) => onToggle(disease, v),
            );
          }).toList(),
        ),
      ],
    );
  }
}

class HiveFramesField extends StatelessWidget {
  final TextEditingController controller;

  const HiveFramesField({super.key, required this.controller});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(labelText: l10n.hiveFrames),
      keyboardType: TextInputType.number,
      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
      textInputAction: TextInputAction.next,
    );
  }
}

final _hiveTypeLabelToKey = {
  for (final e in hiveTypeLabels.entries) e.value: e.key,
};

class HiveTypeDropdown extends StatefulWidget {
  final String value;
  final ValueChanged<String?> onChanged;

  const HiveTypeDropdown({
    super.key,
    required this.value,
    required this.onChanged,
  });

  @override
  State<HiveTypeDropdown> createState() => _HiveTypeDropdownState();
}

class _HiveTypeDropdownState extends State<HiveTypeDropdown> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(
      text: hiveTypeLabels[widget.value] ?? widget.value,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  String _keyFor(String label) => _hiveTypeLabelToKey[label] ?? label;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Autocomplete<String>(
      key: ValueKey(widget.value),
      initialValue: TextEditingValue(
        text: hiveTypeLabels[widget.value] ?? widget.value,
      ),
      optionsBuilder: (textEditingValue) {
        if (textEditingValue.text.isEmpty) {
          return hiveTypeLabels.values;
        }
        final query = textEditingValue.text.toLowerCase();
        return hiveTypeLabels.values.where(
          (label) => label.toLowerCase().contains(query),
        );
      },
      onSelected: (label) {
        widget.onChanged(_keyFor(label));
      },
      fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
        return TextFormField(
          controller: controller,
          focusNode: focusNode,
          decoration: InputDecoration(labelText: l10n.hiveType),
          onChanged: (v) => widget.onChanged(_keyFor(v)),
          validator: (v) =>
              (v == null || v.trim().isEmpty) ? l10n.hiveTypeRequired : null,
        );
      },
    );
  }
}
