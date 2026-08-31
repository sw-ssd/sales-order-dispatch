// This is a generated file - do not edit.
//
// Generated from salesorder/v1/common.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// PrintPaperSize:出貨單/揀貨單列印紙張尺寸。
class PrintPaperSize extends $pb.ProtobufEnum {
  static const PrintPaperSize PRINT_PAPER_SIZE_UNSPECIFIED =
      PrintPaperSize._(0, _omitEnumNames ? '' : 'PRINT_PAPER_SIZE_UNSPECIFIED');
  static const PrintPaperSize PRINT_PAPER_SIZE_A4 =
      PrintPaperSize._(1, _omitEnumNames ? '' : 'PRINT_PAPER_SIZE_A4');
  static const PrintPaperSize PRINT_PAPER_SIZE_A5 =
      PrintPaperSize._(2, _omitEnumNames ? '' : 'PRINT_PAPER_SIZE_A5');
  static const PrintPaperSize PRINT_PAPER_SIZE_80MM =
      PrintPaperSize._(3, _omitEnumNames ? '' : 'PRINT_PAPER_SIZE_80MM');

  static const $core.List<PrintPaperSize> values = <PrintPaperSize>[
    PRINT_PAPER_SIZE_UNSPECIFIED,
    PRINT_PAPER_SIZE_A4,
    PRINT_PAPER_SIZE_A5,
    PRINT_PAPER_SIZE_80MM,
  ];

  static final $core.List<PrintPaperSize?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static PrintPaperSize? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const PrintPaperSize._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
