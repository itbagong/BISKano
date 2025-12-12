/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {StyleSheet, Text, TouchableOpacity, View} from 'react-native';
import React from 'react';
import Dropdown from '@components/dropdown';
import {Datepicker} from '@ui-kitten/components';
import {Colors, Mixins, Typography} from 'utils';
import {Button, Icon} from '@ui-kitten/components';
import {useActions} from '@overmind/index';
import Modal from 'react-native-modal';

type Props = {
  isOpen: any;
  onClose: any;
  navigation: any;
};

const ModalAddNew = (props: Props) => {
  const {getsSite, createNewRoutine} = useActions();
  const [formdata, setFormdata] = React.useState({
    SiteID: '',
    ExecutionDate: new Date(),
  });
  const [dataSites, setDataSites] = React.useState([]);
  const [disabledSite, setDisabledSite] = React.useState(false);

  React.useEffect(() => {
    if (props.isOpen) {
      getsSite({Select: ['_id', 'Name'], Sort: ['_id']}).then(res => {
        setDataSites(res);
        if (res.length === 1) {
          setFormdata({...formdata, SiteID: res[0]._id});
          setDisabledSite(true);
        }
      });
    }
  }, [props.isOpen]);

  const onSubmit = () => {
    createNewRoutine(formdata)
      .then(res => {
        props.navigation.navigate('RoutineDetails', {_id: res._id});
      })
      .finally(() => {
        props.onClose();
      });
  };
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
        <Text style={styles.title}>Add New</Text>
        <View style={styles.closeButton}>
          <TouchableOpacity onPress={() => props.onClose()}>
            <Icon
              name="close-outline"
              style={{width: 30, height: 30}}
              fill={Colors.SHADES.dark[50]}
            />
          </TouchableOpacity>
        </View>
        <View style={{marginTop: Mixins.scaleSize(10)}}>
          <Dropdown
            items={dataSites.map((o: any) => {
              return {value: o._id, label: o.Name};
            })}
            disabled={disabledSite}
            value={formdata.SiteID}
            onSelect={(selected: any) => {
              setFormdata({...formdata, SiteID: selected});
            }}
            inputLabel="Site"
            isRequired
          />
        </View>
        <View style={{marginBottom: Mixins.scaleSize(20)}}>
          <Text style={[styles.label]}>Execution date</Text>
          <Datepicker
            date={formdata.ExecutionDate}
            onSelect={nextDate => {
              setFormdata({
                ...formdata,
                ExecutionDate: nextDate,
              });
            }}
          />
        </View>
        <Button
          onPress={() => onSubmit()}
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
