import { Animated, StyleSheet, Text, View } from 'react-native'
import React from 'react'
import Modal from 'react-native-modal';
import { Mixins, Typography } from 'utils';
import { default as theme } from '../../custom-theme.json';
import * as Progress from 'react-native-progress';

type Props = {
  show: boolean;
}

const Layout = (props: Props) => {
  const [scaleValue] = React.useState(new Animated.Value(1));
  React.useEffect(() => {
    startAnimation();
  }, []);

  const startAnimation = () => {
    Animated.loop(
      Animated.sequence([
        Animated.timing(scaleValue, {
          toValue: 1.2,
          duration: 1000,
          useNativeDriver: true,
        }),
        Animated.timing(scaleValue, {
          toValue: 1,
          duration: 1000,
          useNativeDriver: true,
        }),
      ]),
    ).start();
  };
  return (
    <Modal
      isVisible={props.show}
      backdropColor="#B4B3DB"
      backdropOpacity={0.8}
      animationIn="zoomInDown"
      animationOut="zoomOutUp"
      animationInTiming={600}
      animationOutTiming={600}
      backdropTransitionInTiming={600}
      backdropTransitionOutTiming={600}>
      <View style={styles.modalProgressContainer}>
        <Progress.Bar
          width={Mixins.scaleSize(150)}
          height={Mixins.scaleSize(15)}
          borderRadius={10}
          color={theme['color-primary-600']}
          animationType="spring"
          indeterminate={true}
        />
        <View style={{ marginTop: Mixins.scaleSize(20) }}>
          <Animated.Text
            style={{
              ...Typography.textLgPlusSemiBold,
              transform: [{ scale: scaleValue }],
            }}>
            Loading...
          </Animated.Text>
        </View>
        {/* </View> */}
      </View>
    </Modal>
  )
}

export default Layout

const styles = StyleSheet.create({
  modalProgressContainer: {
    flex: 1,
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
  },
})