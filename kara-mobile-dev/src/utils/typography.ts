import React from 'react';
import {Text, StyleSheet} from 'react-native';
import {scaleFont} from './mixins';

// FONT FAMILY
export const FONT_FAMILY_REGULAR = 'Poppins';
export const FONT_FAMILY_BOLD = 'Poppins Bold';
export const FONT_FAMILY_SEMIBOLD = 'Poppins SemiBold';

// FONT WEIGHT
export const FONT_WEIGHT_REGULAR = '400';
export const FONT_WEIGHT_BOLD = '700';

// FONT SIZE
export const FONT_SIZE_16 = scaleFont(16);
export const FONT_SIZE_14 = scaleFont(14);
export const FONT_SIZE_12 = scaleFont(12);

// LINE HEIGHT
export const LINE_HEIGHT_24 = scaleFont(24);
export const LINE_HEIGHT_20 = scaleFont(20);
export const LINE_HEIGHT_16 = scaleFont(16);

// FONT STYLE
export const FONT_REGULAR = {
  fontFamily: FONT_FAMILY_REGULAR,
  fontWeight: FONT_WEIGHT_REGULAR,
};

export const FONT_BOLD = {
  fontFamily: FONT_FAMILY_BOLD,
  fontWeight: FONT_WEIGHT_BOLD,
};

export const typography = () => {
  let text: any = Text;
  let oldRender = text.render;

  text.render = function (...args: any) {
    let origin: any = oldRender.call(this, ...args);

    return React.cloneElement(origin, {
      style:
        origin.props.style?.fontWeight === 'bold' ||
        origin.props.style?.fontWeight === '700'
          ? [styles.boldText, origin.props.style, {fontWeight: 'normal'}]
          : [styles.defaultText, origin.props.style, {fontWeight: 'normal'}],
    });
  };
};

export const fontStyle = (size: number, weight = 'regular') => {
  let fontFamilyToUse = '';
  switch (weight) {
    case 'regular':
      fontFamilyToUse = FONT_FAMILY_REGULAR;
      break;
    case 'semibold':
      fontFamilyToUse = FONT_FAMILY_SEMIBOLD;
      break;
    case 'bold':
      fontFamilyToUse = FONT_FAMILY_BOLD;
      break;
    default:
      fontFamilyToUse = FONT_FAMILY_REGULAR;
      break;
  }
  // const fontFamilyToUse =
  //   weight === 'regular' ? FONT_FAMILY_REGULAR : FONT_FAMILY_BOLD;
  const style = {
    fontSize: scaleFont(size),
    fontFamily: fontFamilyToUse,
  };
  return style;
};

export const headerLg = fontStyle(34, 'regular');
export const headerLgSemiBold = fontStyle(34, 'semibold');
export const headerLgBold = fontStyle(34, 'bold');
export const headerMd = fontStyle(28, 'regular');
export const headerMdSemiBold = fontStyle(28, 'semibold');
export const headerMdBold = fontStyle(28, 'bold');
export const headerSm = fontStyle(24, 'regular');
export const headerSmSemiBold = fontStyle(24, 'semibold');
export const headerSmBold = fontStyle(24, 'bold');

export const textLgPlus = fontStyle(19, 'regular');
export const textLgPlusSemiBold = fontStyle(19, 'semibold');
export const textLgPlusBold = fontStyle(19, 'bold');
export const textLg = fontStyle(16, 'regular');
export const textLgSemiBold = fontStyle(16, 'semibold');
export const textLgBold = fontStyle(16, 'bold');
export const textMdPlus = fontStyle(14, 'regular');
export const textMdPlusSemiBold = fontStyle(14, 'semibold');
export const textMdPlusBold = fontStyle(14, 'bold');
export const textMd = fontStyle(12, 'regular');
export const textMdSemiBold = fontStyle(12, 'semibold');
export const textMdBold = fontStyle(12, 'bold');
export const textSm = fontStyle(10, 'regular');
export const textSmSemiBold = fontStyle(10, 'semibold');
export const textSmBold = fontStyle(10, 'bold');

const styles = StyleSheet.create({
  defaultText: {
    fontFamily: 'Roboto',
  },
  boldText: {
    fontFamily: 'Roboto Bold',
  },
});
