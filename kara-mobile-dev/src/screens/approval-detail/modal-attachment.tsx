/* eslint-disable react-native/no-inline-styles */
import {
  Linking,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import {Colors, Helper, Mixins, Typography} from 'utils';
import {Icon} from '@ui-kitten/components';
import Modal from 'react-native-modal';
import s from '@components/styles';
import moment from 'moment';

type Props = {
  isOpen: any;
  onClose: any;
  data: any;
};

const ModalAttachment = (props: Props) => {
  const openLink = async (item: any) => {
    const url = Helper.getAssetURL(item._id);
    Linking.openURL(url);
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
        <Text style={styles.title}>Attachment</Text>
        <View style={styles.closeButton}>
          <TouchableOpacity onPress={() => props.onClose()}>
            <Icon
              name="close-outline"
              style={{width: 30, height: 30}}
              fill={Colors.SHADES.dark[50]}
            />
          </TouchableOpacity>
        </View>

        <View style={{marginVertical: Mixins.scaleSize(20), flex: 1}}>
          <ScrollView style={{gap: 5}}>
            {props.data?.map((item: any, i: number) => (
              <View key={i} style={styles.card}>
                <View style={s.row}>
                  <Text
                    style={{
                      flex: 1,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    File name
                  </Text>
                  <TouchableOpacity
                    style={{flex: 1.5}}
                    onPress={() => openLink(item)}>
                    <Text
                      style={{
                        ...Typography.textMdPlus,
                        color: Colors.BLACK,
                      }}>
                      :{' '}
                      <Text
                        style={{
                          ...Typography.textMdPlus,
                          color: Colors.SECONDARY.blue,
                        }}>
                        {item.Data.FileName}
                      </Text>
                    </Text>
                  </TouchableOpacity>
                </View>
                <View style={s.row}>
                  <Text
                    style={{
                      flex: 1,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    Description
                  </Text>
                  <Text
                    style={{
                      flex: 1.5,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    : {item.Data.Description}
                  </Text>
                </View>
                <View style={s.row}>
                  <Text
                    style={{
                      flex: 1,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    Upload date
                  </Text>
                  <Text
                    style={{
                      flex: 1.5,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    :{' '}
                    {moment(item.Data.UploadDate).format(
                      'DD-MMM-YYYY hh:mm:ss',
                    )}
                  </Text>
                </View>
              </View>
            ))}
          </ScrollView>
        </View>
      </View>
    </Modal>
  );
};

export default ModalAttachment;

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.WHITE,
    padding: Mixins.scaleSize(20),
    borderRadius: Mixins.scaleSize(15),
    position: 'relative',
    flex: 1,
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
  card: {
    borderWidth: 1,
    borderColor: Colors.SHADES.red[300],
    borderRadius: 10,
    padding: Mixins.scaleSize(5),
    paddingHorizontal: Mixins.scaleSize(10),
    marginBottom: Mixins.scaleSize(10),
  },
});
