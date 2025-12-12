/* eslint-disable react-hooks/exhaustive-deps */
import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  StyleProp,
  ViewStyle,
} from 'react-native';
import {Colors, Mixins, Typography} from '@utils/index';

type Props = {
  containerStyle?: StyleProp<ViewStyle> | {};
  dataCount: any;
  showPer: any;
  state: any;
  onChangeState: (value: number) => void;
};

const Pagination = (props: Props) => {
  const [allPages, setAllPage] = React.useState([] as any);

  const onChangeActivePage = (val: any) => {
    props.onChangeState(val);
  };
  const initPaging = () => {
    let remainder = props.dataCount % props.showPer;
    let count = props.dataCount / props.showPer;
    if (remainder > 0) {
      count = Math.ceil(count);
    }
    let pages = Array.from(Array(count + 1).keys()).slice(1);
    // console.log('page count: ', count, 'pgaes: ', pages);
    setAllPage([...pages]);
    // if (count <= 5) {
    //   setShowPage([...pages]);
    //   setIsMorePage(false);
    // } else {
    //   setIsMorePage(true);
    // }
  };
  React.useEffect(() => {
    // console.log(props.dataCount);
    if (props.dataCount) {
      initPaging();
    }
    return () => {};
  }, [props.showPer, props.state]);
  // React.useEffect(() => {
  //   return () => {};
  // }, [props.state]);
  return (
    <View style={[styles.container, props.containerStyle]}>
      <TouchableOpacity
        style={[styles.button, props.state === 1 ? styles.buttonDisabled : {}]}
        disabled={props.state === 1}
        onPress={() => {
          if (props.state !== 1) {
            onChangeActivePage(props.state - 1);
          }
        }}>
        <Text
          style={[
            styles.labelButton,
            props.state === 1 ? styles.buttonDisabled : {},
          ]}>
          Previous
        </Text>
      </TouchableOpacity>
      <View style={styles.pagingNumberBody}>
        <Text style={styles.label}>{props.state}</Text>
      </View>
      {/* <View style={styles.pagingNumberBody}>
        <TouchableOpacity
          style={styles.buttonPrev}
          disabled={props.state === 1}
          onPress={() => {
            if (props.state !== 1) {
              onChangeActivePage(props.state - 1);
            }
            console.log(props.state);
          }}>
          <FontAwesome
            name="caret-left"
            color={Colors.PRIMARY.green}
            size={20}
          />
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.buttonNext}
          onPress={() => {
            if (props.state !== allPages[allPages.length - 1]) {
              onChangeActivePage(props.state + 1);
            }
          }}>
          <FontAwesome
            name="caret-right"
            color={Colors.PRIMARY.green}
            size={20}
          />
        </TouchableOpacity>
      </View> */}
      <TouchableOpacity
        style={[
          styles.button,
          props.state === allPages[allPages.length - 1]
            ? styles.buttonDisabled
            : {},
        ]}
        disabled={props.state === allPages[allPages.length - 1]}
        onPress={() => {
          if (props.state !== allPages[allPages.length - 1]) {
            onChangeActivePage(props.state + 1);
          }
        }}>
        <Text
          style={[
            styles.labelButton,
            props.state === allPages[allPages.length - 1]
              ? styles.buttonDisabled
              : {},
          ]}>
          Next
        </Text>
      </TouchableOpacity>
    </View>
  );
};
const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 10,
  },
  button: {
    flex: 1,
    borderWidth: 1,
    borderColor: Colors.SHADES.red[400],
    backgroundColor: Colors.SHADES.red[400],
    borderRadius: Mixins.scaleSize(5),
    paddingHorizontal: Mixins.scaleSize(10),
    paddingVertical: Mixins.scaleSize(5),
  },
  buttonPage: {
    borderTopWidth: 1,
    borderBottomWidth: 1,
    borderColor: Colors.SHADES.red[400],
    paddingHorizontal: Mixins.scaleSize(10),
    paddingVertical: Mixins.scaleSize(5),
  },
  notLastPage: {
    borderRightWidth: 1,
    borderColor: Colors.SHADES.red[400],
  },
  activePage: {
    backgroundColor: Colors.SHADES.red[400],
  },
  txtActivePage: {
    color: Colors.WHITE,
  },
  label: {
    ...Typography.textMdPlusSemiBold,
    color: Colors.SHADES.red[400],
    textAlign: 'center',
  },
  labelButton: {
    ...Typography.textMdPlusSemiBold,
    color: Colors.WHITE,
    textAlign: 'center',
  },
  labelDisabled: {
    color: Colors.SHADES.gray[800],
  },
  pagingNumberBody: {
    borderColor: Colors.PRIMARY.red,
    borderWidth: 1,
    borderRadius: Mixins.scaleSize(5),
    paddingHorizontal: Mixins.scaleSize(20),
    paddingVertical: Mixins.scaleSize(5),
  },
  buttonDisabled: {
    backgroundColor: Colors.SHADES.gray[200],
    borderColor: Colors.SHADES.gray[300],
    color: Colors.SHADES.gray[800],
  },
});

export default Pagination;
