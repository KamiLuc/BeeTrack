import '../../../core/api/api_client.dart';
import 'invitation_model.dart';

class InvitationRepository {
  final ApiClient _api;

  const InvitationRepository({required ApiClient api}) : _api = api;

  Future<void> sendInvitation(int apiaryId, String email) async {
    await _api.dio.post(
      '/api/v1/apiaries/$apiaryId/invitations',
      data: {'email': email},
    );
  }

  Future<ApiaryMembersData> listForApiary(int apiaryId) async {
    final response = await _api.dio.get('/api/v1/apiaries/$apiaryId/invitations');
    final data = response.data as Map<String, dynamic>;
    return ApiaryMembersData(
      members: (data['members'] as List)
          .map((e) => ApiaryMemberInfo.fromJson(e as Map<String, dynamic>))
          .toList(),
      invitations: (data['invitations'] as List)
          .map((e) => ApiaryInvitation.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<void> cancelInvitation(int apiaryId, int invitationId) async {
    await _api.dio.delete('/api/v1/apiaries/$apiaryId/invitations/$invitationId');
  }

  Future<void> removeMember(int apiaryId, int userId) async {
    await _api.dio.delete('/api/v1/apiaries/$apiaryId/members/$userId');
  }

  Future<void> leaveApiary(int apiaryId) async {
    await _api.dio.delete('/api/v1/apiaries/$apiaryId/leave');
  }

  Future<List<MyInvitation>> listMine() async {
    final response = await _api.dio.get('/api/v1/invitations');
    return (response.data as List)
        .map((e) => MyInvitation.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<int> countMine() async {
    final response = await _api.dio.get('/api/v1/invitations/count');
    return (response.data as Map<String, dynamic>)['count'] as int;
  }

  Future<void> accept(int invitationId) async {
    await _api.dio.post('/api/v1/invitations/$invitationId/accept');
  }

  Future<void> decline(int invitationId) async {
    await _api.dio.post('/api/v1/invitations/$invitationId/decline');
  }
}
