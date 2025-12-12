/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {StyleSheet, Text, TouchableOpacity, View} from 'react-native';
import React from 'react';
import {Input} from '@ui-kitten/components';
import {Colors, Mixins, Typography} from 'utils';
import {Button, Icon} from '@ui-kitten/components';
import Modal from 'react-native-modal';

type Props = {
  isOpen: any;
  onClose: any;
  onSubmit: any;
  Op: any;
};

const ModalAddNew = (props: Props) => {
  const [message, setMessage] = React.useState('');

  React.useEffect(() => {
    if (props.isOpen) {
      setMessage('');
    }
  }, [props.isOpen]);
  return (
    <Modal
      testID={'modal'}
      onBackdropPress={props.onClose}
      isVisible={props.isOpen}
      backdropColor="#B4B3DB"
      backdropOpacity={0.8}
      animationIn="zoomInDown"
      animationOut="zoomOutUp"
      animationInTiming={600}
      animationOutTiming={600}
      backdropTransitionInTiming={600}
      backdropTransitionOutTiming={600}>
      <View style={styles.container}>
        <Text style={styles.title}>{props.Op} Message</Text>
        <View style={styles.closeButton}>
          <TouchableOpacity onPress={() => props.onClose()}>
            <Icon
              name="close-outline"
              style={{width: 30, height: 30}}
              fill={Colors.SHADES.dark[50]}
            />
          </TouchableOpacity>
        </View>
        <View style={{marginVertical: Mixins.scaleSize(20)}}>
          <Input
            value={message}
            multiline
            numberOfLines={4}
            textStyle={{paddingHorizontal: 0}}
            placeholder="Text"
            onChangeText={nextValue => {
              setMessage(nextValue);
            }}
          />
        </View>
        <Button
          onPress={() => props.onSubmit(message)}
          style={styles.buttonSubmit}
          size="large"
          status="primary">
          {() => (
            <Text
              style={{
                ...Typography.textLgSemiBold,
                color: 'white',
                marginLeft: Mixins.scaleSize(10),
              }}>
              Submit
            </Text>
          )}
        </Button>
      </View>
    </Modal>
  );
};

export default ModalAddNew;

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.WHITE,
    padding: Mixins.scaleSize(20),
    borderRadius: Mixins.scaleSize(15),
    position: 'relative',
  },
  title: {
    ...Typography.textLgPlusSemiBold,
    textAlign: 'center',
    color: Colors.BLACK,
  },
  label: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
  buttonSubmit: {
    borderRadius: Mixins.scaleSize(8),
    alignItems: 'center',
    justifyContent: 'center',
  },
  buttonAction: {
    flex: 1,
    flexDirection: 'row',
    padding: Mixins.scaleSize(10),
    borderRadius: Mixins.scaleSize(5),
    justifyContent: 'center',
    alignItems: 'center',
    gap: Mixins.scaleSize(10),
  },
  labelAction: {
    ...Typography.textMdPlusSemiBold,
  },
  closeButton: {
    position: 'absolute',
    top: 20,
    right: 20,
  },
});
