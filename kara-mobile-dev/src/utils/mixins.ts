import {Dimensions, PixelRatio} from 'react-native';

export const WINDOW_WIDTH = Dimensions.get('window').width;
export const WINDOW_HEIGHT = Dimensions.get('window').height;
const guidelineBaseWidth = 375;

export const scaleSize = (size: number) =>
  (WINDOW_WIDTH / guidelineBaseWidth) * size;

export const scaleFont = (size: number) => {
  const scale = PixelRatio.getFontScale();
  if (scale < 1) return size * (1 - (scale - 1));
  else return size * (1 + (1 - scale));
};

function dimensions(
  top: number,
  right = top,
  bottom = top,
  left = right,
  property: string,
) {
  let styles = {} as any;

  styles[`${property}Top`] = top;
  styles[`${property}Right`] = right;
  styles[`${property}Bottom`] = bottom;
  styles[`${property}Left`] = left;

  return styles;
}

export function getColor(value: number, target: number, achieved = false) {
  let color = '';
  if (achieved) color = '#3b8b72';
  else {
    let diff = Math.abs(value - target);
    let ratio = diff / target;

    if (ratio <= 0.05) color = '#e77836';
    if (ratio <= 0.15 && ratio >= 0.05) color = '#a8a277';
    else if (ratio <= 0.15) color = '#3b8b72';
    else color = '#b41313';
  }
  return color;
}

export function margin(
  top: number,
  right: number,
  bottom: number,
  left: number,
) {
  return dimensions(top, right, bottom, left, 'margin');
}

export function padding(
  top: number,
  right: number,
  bottom: number,
  left: number,
) {
  return dimensions(top, right, bottom, left, 'padding');
}

export function boxShadow(
  color: string,
  offset = {height: 2, width: 2},
  radius = 8,
  opacity = 0.2,
) {
  return {
    shadowColor: color,
    shadowOffset: offset,
    shadowOpacity: opacity,
    shadowRadius: radius,
    elevation: radius,
  };
}
