import React from 'react';
import {StyleProp, StyleSheet, View, ViewStyle} from 'react-native';

import {Mixins, Colors} from '@utils/index';

interface Props {
  containerStyle?: StyleProp<ViewStyle> | {};
  innerStyle?: StyleProp<ViewStyle> | {};
  children?: any;
}
const Card: React.FC<Props> = (props: any) => {
  return (
    <View style={[styles.column, styles.card, props.containerStyle]}>
      {props?.children && (
        <View style={props.innerStyle}>{props?.children}</View>
      )}
    </View>
  );
};
const styles = StyleSheet.create({
  column: {
    flexDirection: 'column',
  },
  card: {
    backgroundColor: Colors.WHITE,
    borderRadius: Mixins.scaleSize(10),
    borderColor: Colors.SHADES.dark[50],
    borderWidth: Mixins.scaleSize(1),
    elevation: 1,
    shadowRadius: 5,
    shadowColor: Colors.SHADES.gray[200],
    shadowOffset: {width: 0, height: 0},
    shadowOpacity: 70,
  },
});
export default Card;
