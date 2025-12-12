import {StyleSheet} from 'react-native';
import {Mixins} from 'utils';

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingVertical: Mixins.scaleSize(10),
    paddingHorizontal: Mixins.scaleSize(14),
  },
  row: {
    flexDirection: 'row',
  },
  column: {
    flexDirection: 'column',
  },
  bold: {
    fontFamily: 'Poppins Bold',
  },
  semibold: {
    fontFamily: 'Poppins SemiBold',
  },
  regular: {
    fontFamily: 'Poppins',
  },
  floatLeft: {
    flex: 1,
    alignItems: 'flex-start',
  },
  floatRight: {
    flex: 1,
    alignItems: 'flex-end',
  },
});

export default styles;
