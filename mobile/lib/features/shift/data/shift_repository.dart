import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../core/network/api_exception.dart';
import '../../../core/providers.dart';
import '../models/shift.dart';

class ShiftRepository {
  ShiftRepository(this._api);
  final ApiClient _api;

  /// The cashier's open shift, or null when no shift is open.
  Future<Shift?> current() async {
    try {
      return await _api.get('/shifts/current', Shift.fromJson);
    } on ApiException catch (e) {
      if (e.status == 404) return null;
      rethrow;
    }
  }

  Future<Shift> open(int openingCash, {String? note}) async {
    final res = await _api.post('/shifts/open', body: {'opening_cash': openingCash, 'note': note});
    return Shift.fromJson(res['data'] as Map<String, dynamic>);
  }

  Future<Shift> close(int actualCash, {String? note}) async {
    final res = await _api.post('/shifts/current/close', body: {'actual_cash': actualCash, 'note': note});
    return Shift.fromJson(res['data'] as Map<String, dynamic>);
  }
}

final shiftRepositoryProvider = Provider((ref) => ShiftRepository(ref.watch(apiClientProvider)));
