import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/shift_repository.dart';
import '../models/shift.dart';

/// The signed-in cashier's open shift (null when none is open).
class CurrentShiftController extends AsyncNotifier<Shift?> {
  @override
  Future<Shift?> build() => ref.read(shiftRepositoryProvider).current();

  Future<void> open(int openingCash, {String? note}) async {
    await ref.read(shiftRepositoryProvider).open(openingCash, note: note);
    state = AsyncData(await ref.read(shiftRepositoryProvider).current());
  }

  /// Closes the shift and returns the final recap.
  Future<Shift> close(int actualCash, {String? note}) async {
    final closed = await ref.read(shiftRepositoryProvider).close(actualCash, note: note);
    state = const AsyncData(null);
    return closed;
  }

  Future<void> refresh() async {
    state = AsyncData(await ref.read(shiftRepositoryProvider).current());
  }
}

final currentShiftProvider = AsyncNotifierProvider<CurrentShiftController, Shift?>(CurrentShiftController.new);
