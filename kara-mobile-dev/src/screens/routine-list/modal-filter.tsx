/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {StyleSheet, Text, TouchableOpacity, View} from 'react-native';
import React from 'react';
import Modal from 'react-native-modal';
import {Colors, Mixins, Typography} from 'utils';
import {Button, Icon, Input} from '@ui-kitten/components';
import {useActions} from '@overmind/index';
import Dropdown from '@components/dropdown';
import {Datepicker} from '@ui-kitten/components';
import {default as theme} from '../../custom-theme.json';

type Props = {
  onApply: any;
  onReset: any;
  isOpen: any;
  onClose: any;
};

const ModalFilter = (props: Props) => {
  const {getsSite} = useActions();
  const [formdata, setFormdata] = React.useState({
    keyword: '',
    site: '',
    executionDate: {
      from: null,
      to: null,
    },
  });
  const [dataSites, setDataSites] = React.useState([]);
  React.useEffect(() => {
    // if (props.isOpen) {
    getsSite({Select: ['_id', 'Name'], Sort: ['_id']}).then(res => {
      setDataSites(res);
      if (res.length === 1) {
        setFormdata({...formdata, site: res[0]._id});
      }
    });
    // }
  }, []);
  return (
    <Modal
      testID={'modal'}
      onBackdropPress={props.onClose}
      isVisible={props.isOpen}
      onSwipeComplete={props.onClose}
      backdropColor="#B4B3DB"
      swipeDirection={['up', 'left', 'right', 'down']}
      style={styles.view}>
      <View style={styles.container}>
        <Text style={styles.title}>Filters</Text>
        <View style={{marginTop: Mixins.scaleSize(10)}}>
          <View style={{marginBottom: Mixins.scaleSize(10)}}>
            <Text style={[styles.label]}>Search</Text>
            <Input
              value={formdata.keyword}
              placeholder="search..."
              accessoryLeft={
                <Icon
                  name={'search-outline'}
                  fill={Colors.SHADES.gray[800]}
                  style={{
                    width: Mixins.scaleSize(15),
                    height: Mixins.scaleSize(15),
                  }}
                />
              }
              onChangeText={nextValue =>
                setFormdata({...formdata, keyword: nextValue})
              }
            />
          </View>
          <Dropdown
            items={dataSites.map((o: any) => {
              return {value: o._id, label: o.Name};
            })}
            value={formdata.site}
            onSelect={(selected: any) => {
              setFormdata({...formdata, site: selected});
            }}
            inputLabel="Site"
            isRequired
          />
        </View>
        <View style={{marginBottom: Mixins.scaleSize(20)}}>
          <Text style={[styles.label]}>Execution date</Text>
          <View
            style={{
              flexDirection: 'row',
              gap: Mixins.scaleSize(10),
              alignItems: 'center',
            }}>
            <Datepicker
              date={formdata.executionDate.from}
              style={{flex: 1}}
              onSelect={nextDate => {
                setFormdata({
                  ...formdata,
                  executionDate: {...formdata.executionDate, from: nextDate},
                });
              }}
            />
            <Text>to</Text>
            <Datepicker
              date={formdata.executionDate.to}
              style={{flex: 1}}
              onSelect={nextDate =>
                setFormdata({
                  ...formdata,
                  executionDate: {...formdata.executionDate, to: nextDate},
                })
              }
            />
          </View>
        </View>
        <View
          style={{
            flexDirection: 'row',
            gap: Mixins.scaleSize(10),
            alignItems: 'center',
          }}>
          <View style={{flex: 2}}>
            <Button
              onPress={() => props.onApply(formdata)}
              style={styles.buttonApply}
              size="large"
              status="primary">
              {() => (
                <Text
                  style={{
                    ...Typography.textLgSemiBold,
                    color: 'white',
                    marginLeft: Mixins.scaleSize(10),
                  }}>
                  Apply filter
                </Text>
              )}
            </Button>
          </View>
          <View style={{flex: 1}}>
            <TouchableOpacity
              onPress={() => {
                props.onReset();
                setFormdata({
                  keyword: '',
                  site: '',
                  executionDate: {
                    from: null,
                    to: null,
                  },
                });
                props.onClose();
              }}
              style={{
                ...styles.buttonAction,
                backgroundColor: theme['color-primary-300'],
              }}>
              <Text
                style={{
                  ...styles.labelAction,
                  color: theme['color-primary-900'],
                }}>
                Reset
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      </View>
    </Modal>
  );
};

export default ModalFilter;

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.WHITE,
    padding: Mixins.scaleSize(14),
    borderTopLeftRadius: Mixins.scaleSize(40),
    borderTopRightRadius: Mixins.scaleSize(40),
  },
  view: {
    justifyContent: 'flex-end',
    margin: 0,
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
  buttonApply: {
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
});
