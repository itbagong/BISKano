/* eslint-disable react-native/no-inline-styles */
/* eslint-disable @typescript-eslint/no-unused-vars */
import React, {useEffect, useState} from 'react';
import {StyleSheet, View, Text} from 'react-native';
import {Dropdown} from 'react-native-element-dropdown';
import {Icon, Input} from '@ui-kitten/components';
import {Colors, Mixins, Typography} from 'utils';

type Props = {
  items: any;
  value: any;
  onSelect: (selected: any) => void;
  inputLabel?: string;
  onSearch?: (text: string) => void;
  disabled?: boolean;
  isRequired?: boolean;
  error?: boolean;
};
const DropdownComponent = (props: Props) => {
  const [data, setData] = useState([]);
  const [value, setValue] = useState(null);

  useEffect(() => {
    setValue(props.value);
    return () => {};
  }, [props.value]);

  useEffect(() => {
    setData(props.items);
    return () => {};
  }, [props.items]);
  const renderItem = (item: any) => {
    return (
      <View style={styles.item}>
        <Text style={styles.textItem}>{item.label}</Text>
        {item.value === value && (
          <Icon
            style={styles.icon}
            fill="black"
            name="checkmark-square-outline"
          />
        )}
      </View>
    );
  };

  return (
    <View style={styles.dropdownContainer}>
      {props.inputLabel !== '' && (
        <Text style={[styles.label]}>{props.inputLabel}</Text>
      )}
      <Dropdown
        style={{
          ...styles.dropdown,
          backgroundColor: props.disabled ? '#D0D4DC61' : '#f7f9fc',
          borderColor: !props.error ? '#e4e9f2' : Colors.PRIMARY.red,
        }}
        disable={props.disabled}
        placeholderStyle={styles.placeholderStyle}
        selectedTextStyle={styles.selectedTextStyle}
        inputSearchStyle={styles.inputSearchStyle}
        iconStyle={styles.iconStyle}
        data={data}
        search
        maxHeight={300}
        labelField="label"
        valueField="value"
        placeholder="Select item"
        searchPlaceholder="Search..."
        value={value}
        onChange={(item: any) => {
          props.onSelect(item.value);
        }}
        renderItem={renderItem}
        renderInputSearch={(onSearch: any) => (
          <Input
            style={{padding: Mixins.scaleSize(10)}}
            placeholder="Search"
            onChangeText={nextValue => {
              if (props.onSearch) {
                props.onSearch(nextValue);
              } else {
                onSearch(nextValue);
              }
            }}
          />
        )}
      />
      {props.error && props.isRequired && (
        <Text style={[styles.labelRequired]}>required</Text>
      )}
    </View>
  );
};

export default DropdownComponent;

const styles = StyleSheet.create({
  dropdownContainer: {
    marginBottom: Mixins.scaleSize(10),
  },
  label: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
    marginBottom: Mixins.scaleSize(-3),
  },
  labelRequired: {
    ...Typography.textMd,
    color: Colors.PRIMARY.red,
  },
  dropdown: {
    marginTop: Mixins.scaleSize(5),
    height: 50,
    backgroundColor: '#f7f9fc', //'white',
    borderRadius: Mixins.scaleSize(5),
    paddingHorizontal: Mixins.scaleSize(10),
    borderWidth: 1,
    borderColor: Colors.SHADES.dark[100],
  },
  icon: {
    marginRight: 5,
    width: 25,
    height: 25,
  },
  item: {
    padding: 17,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  textItem: {
    flex: 1,
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
  placeholderStyle: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
  selectedTextStyle: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
  iconStyle: {
    width: 20,
    height: 20,
  },
  inputSearchStyle: {
    height: 40,
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
});
