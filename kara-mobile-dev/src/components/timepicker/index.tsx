/* eslint-disable react-native/no-inline-styles */
import {
  StyleProp,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  ViewStyle,
} from 'react-native';
import React from 'react';
import DateTimePickerModal from 'react-native-modal-datetime-picker';
import {Colors, Mixins, Typography} from 'utils';
import moment from 'moment';

type Props = {
  value: any;
  defaultDate: any;
  disabled?: boolean;
  label?: string;
  containerStyle?: StyleProp<ViewStyle>;
  onChange: (date: Date) => void;
};

const TimePicker = (props: Props) => {
  const [show, setShow] = React.useState(false);
  const setValueDateTime = () => {
    return new Date(
      moment(props.defaultDate).format('YYYY-MM-DD') + ' ' + props.value,
    );
  };
  return (
    <View style={props.containerStyle}>
      {props.label && <Text style={[styles.label]}>{props.label}</Text>}
      <TouchableOpacity
        style={styles.button}
        disabled={props.disabled}
        onPress={() => setShow(true)}>
        <Text
          style={{...Typography.textMdPlus, color: Colors.SHADES.dark[400]}}>
          {props.value ? props.value : '--:--'}
        </Text>
      </TouchableOpacity>
      <DateTimePickerModal
        isVisible={show}
        date={props.value ? setValueDateTime() : new Date()}
        style={{borderRadius: Mixins.scaleSize(10)}}
        mode="time"
        is24Hour
        onConfirm={date => {
          props.onChange(date);
          setShow(false);
        }}
        onCancel={() => setShow(false)}
      />
    </View>
  );
};

export default TimePicker;

const styles = StyleSheet.create({
  button: {
    flex: 1,
    backgroundColor: '#f7f9fc',
    borderColor: '#e4e9f2',
    borderWidth: 1,
    borderRadius: 4,
    paddingVertical: 8,
    paddingHorizontal: 14,
  },
  label: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
});
