/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import {Datepicker, Input} from '@ui-kitten/components';
import {Colors, Mixins, Typography} from 'utils';
import {Button, Icon} from '@ui-kitten/components';
import {useActions} from '@overmind/index';
import Modal from 'react-native-modal';
// import {default as theme} from '../../custom-theme.json';
import {launchCamera} from 'react-native-image-picker';
import {ALERT_TYPE, Toast} from 'react-native-alert-notification';

type Props = {
  isOpen: any;
  onClose: any;
  data: any;
  onSubmit: any;
  refID: string;
};

const ModalAddNew = (props: Props) => {
  const {} = useActions();
  const [asset, setAsset] = React.useState({
    FileName: '',
    Asset: {},
    Content: '',
    ImageUri: '',
    Description: '',
  } as any);
  const [errorForm, setErrorForm] = React.useState({
    Photo: false,
  });
  React.useEffect(() => {
    if (props.isOpen) {
      setAsset({
        FileName: '',
        Asset: {},
        Content: '',
        Description: props.data.Description,
      });
    }
  }, [props.isOpen]);
  const onTakePicture = async () => {
    try {
      setErrorForm({...errorForm, Photo: false});
      launchCamera(
        {
          mediaType: 'photo',
          cameraType: 'back',
          saveToPhotos: true,
          includeBase64: true,
        },
        (response: any) => {
          if (!response.didCancel === true) {
            const file = response.assets[0];
            const _asset = {
              Asset: {
                ContentType: file.type,
                Data: {
                  UploadDate: new Date(),
                  Content: file.base64,
                  FileName: file.fileName,
                  Description: props.data.Description,
                },
                Kind: 'P2H',
                RefID: props.refID,
                OriginalFileName: file.fileName,
              },
              Content: file.base64,
            };
            setAsset({
              ..._asset,
              ImageUri: file.uri,
              FileName: file.fileName,
              Description: props.data.Description,
            });
          }
        },
      );
    } catch (e) {
      handleErrorFilePicker(e);
    }
  };
  const handleErrorFilePicker = (err: unknown) => {
    throw err;
  };
  const onSubmit = () => {
    if (asset.FileName === '') {
      setErrorForm({...errorForm, Photo: true});
      return Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody: 'Photo can not be empty',
      });
    }
    props.onSubmit(asset);
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
        <Text style={styles.title}>Upload {props.data.Title}</Text>
        <View style={styles.closeButton}>
          <TouchableOpacity onPress={() => props.onClose()}>
            <Icon
              name="close-outline"
              style={{width: 30, height: 30}}
              fill={Colors.SHADES.dark[50]}
            />
          </TouchableOpacity>
        </View>
        <ScrollView>
          <View style={{marginTop: Mixins.scaleSize(10)}}>
            <Text style={[styles.label]}>File Name</Text>
            <Input
              disabled
              value={asset.FileName}
              textStyle={{paddingHorizontal: 0, color: Colors.BLACK}}
              placeholder=""
            />
          </View>
          <View style={{marginTop: Mixins.scaleSize(10)}}>
            <Text style={[styles.label]}>Type</Text>
            <Input
              disabled
              value={props.data.Description}
              textStyle={{paddingHorizontal: 0, color: Colors.BLACK}}
              placeholder="Description"
            />
          </View>
          <View style={{marginTop: Mixins.scaleSize(10)}}>
            <Text style={[styles.label]}>Upload date</Text>
            <Datepicker disabled date={new Date()} />
          </View>
          <View style={{marginVertical: Mixins.scaleSize(10)}}>
            <View
              style={{
                marginTop: Mixins.scaleSize(10),
                marginBottom: Mixins.scaleSize(20),
              }}>
              <View style={{alignContent: 'center'}}>
                <TouchableOpacity
                  onPress={() => onTakePicture()}
                  style={styles.boxFile}>
                  {asset.FileName ? (
                    <Image
                      source={{uri: asset.ImageUri}}
                      style={{
                        width: '100%',
                        height: '100%',
                        borderRadius: Mixins.scaleSize(8),
                      }}
                      resizeMode="contain"
                    />
                  ) : (
                    <Icon
                      style={{
                        width: Mixins.scaleSize(90),
                        height: Mixins.scaleSize(90),
                      }}
                      fill={Colors.SHADES.purple[500]}
                      name="camera-outline"
                    />
                  )}
                  {asset.FileName && (
                    <TouchableOpacity
                      style={styles.closebutton}
                      onPress={() => {
                        setAsset({
                          FileName: '',
                          Asset: {},
                          Content: '',
                          ImageUri: '',
                          Descriptopn: '',
                        });
                      }}>
                      <Icon
                        style={{
                          width: Mixins.scaleSize(25),
                          height: Mixins.scaleSize(25),
                        }}
                        fill={Colors.SHADES.dark[700]}
                        name="close-circle"
                      />
                    </TouchableOpacity>
                  )}
                </TouchableOpacity>
              </View>
              {errorForm.Photo && (
                <Text
                  style={{
                    ...Typography.textMd,
                    color: Colors.PRIMARY.red,
                  }}>
                  photo is required
                </Text>
              )}
            </View>
          </View>
        </ScrollView>
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
    flex: 1,
    backgroundColor: Colors.WHITE,
    padding: Mixins.scaleSize(20),
    borderRadius: Mixins.scaleSize(15),
    position: 'relative',
  },
  title: {
    ...Typography.textLgPlusSemiBold,
    color: Colors.BLACK,
  },
  label: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
  buttonSubmit: {
    marginTop: Mixins.scaleSize(20),
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
  boxFile: {
    borderRadius: Mixins.scaleSize(8),
    borderWidth: 2,
    borderStyle: 'dashed',
    borderColor: Colors.SHADES.purple[300],
    backgroundColor: Colors.SHADES.purple[100],
    marginTop: Mixins.scaleSize(10),
    justifyContent: 'center',
    alignItems: 'center',
    width: '100%',
    height: Mixins.scaleSize(150),
    position: 'relative',
  },
  closebuttonFile: {
    position: 'absolute',
    zIndex: 1,
    right: Mixins.scaleSize(5),
    backgroundColor: Colors.WHITE,
    borderRadius: 50,
    top: Mixins.scaleSize(-10),
  },
  closebutton: {
    position: 'absolute',
    zIndex: 1,
    right: Mixins.scaleSize(-2),
    backgroundColor: Colors.WHITE,
    borderRadius: 50,
    top: Mixins.scaleSize(-10),
  },
});
