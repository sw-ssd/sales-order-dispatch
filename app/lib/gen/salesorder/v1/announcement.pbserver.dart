// This is a generated file - do not edit.
//
// Generated from salesorder/v1/announcement.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'announcement.pb.dart' as $0;
import 'announcement.pbjson.dart';

export 'announcement.pb.dart';

abstract class AnnouncementServiceBase extends $pb.GeneratedService {
  $async.Future<$0.ListAnnouncementsResponse> listAnnouncements(
      $pb.ServerContext ctx, $0.ListAnnouncementsRequest request);
  $async.Future<$0.ListActiveAnnouncementsResponse> listActiveAnnouncements(
      $pb.ServerContext ctx, $0.ListActiveAnnouncementsRequest request);
  $async.Future<$0.CreateAnnouncementResponse> createAnnouncement(
      $pb.ServerContext ctx, $0.CreateAnnouncementRequest request);
  $async.Future<$0.UpdateAnnouncementResponse> updateAnnouncement(
      $pb.ServerContext ctx, $0.UpdateAnnouncementRequest request);
  $async.Future<$0.DeleteAnnouncementResponse> deleteAnnouncement(
      $pb.ServerContext ctx, $0.DeleteAnnouncementRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListAnnouncements':
        return $0.ListAnnouncementsRequest();
      case 'ListActiveAnnouncements':
        return $0.ListActiveAnnouncementsRequest();
      case 'CreateAnnouncement':
        return $0.CreateAnnouncementRequest();
      case 'UpdateAnnouncement':
        return $0.UpdateAnnouncementRequest();
      case 'DeleteAnnouncement':
        return $0.DeleteAnnouncementRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListAnnouncements':
        return listAnnouncements(ctx, request as $0.ListAnnouncementsRequest);
      case 'ListActiveAnnouncements':
        return listActiveAnnouncements(
            ctx, request as $0.ListActiveAnnouncementsRequest);
      case 'CreateAnnouncement':
        return createAnnouncement(ctx, request as $0.CreateAnnouncementRequest);
      case 'UpdateAnnouncement':
        return updateAnnouncement(ctx, request as $0.UpdateAnnouncementRequest);
      case 'DeleteAnnouncement':
        return deleteAnnouncement(ctx, request as $0.DeleteAnnouncementRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      AnnouncementServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => AnnouncementServiceBase$messageJson;
}
