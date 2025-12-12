/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {
  StatusBar,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  Image,
  ScrollView,
  PermissionsAndroid,
  Animated,
} from 'react-native';
import React from 'react';
import {useActions, useAppState} from '@overmind/index';
import {Button, Icon} from '@ui-kitten/components';
import {Colors, Mixins, Typography} from 'utils';
import {default as theme} from '../../custom-theme.json';
// import moment from 'moment';
import Modal from 'react-native-modal';
import Dropdown from '@components/dropdown';
import {ALERT_TYPE, Toast} from 'react-native-alert-notification';
import {launchCamera} from 'react-native-image-picker';
import Geolocation from 'react-native-geolocation-service';
import {useIsFocused} from '@react-navigation/native';
import container, {ContainerContext} from '@components/container';
import DateLocInformation from './date-loc-information';
import CardAttendanceInfo from './card-attendance-information';
import AttendanceHistory from './attendace-history';
import images from 'assets/images';
import * as Progress from 'react-native-progress';

type Props = {
  navigation: any;
};

const Layout = (props: Props) => {
  const isFocused = useIsFocused();
  const ctx = React.useContext(ContainerContext);

  const {} = props;
  // const route = useRoute();
  const {
    checkIn,
    checkOut,
    getNearestLocation,
    getMasterBus,
    getStatus,
    writeBatchWithContent,
    deleteAttendance,
  } = useActions();
  const {isCheckIn} = useAppState();
  const [keyStatusBar, setKeyStatusBar] = React.useState(0);
  const [keyAttendaceHistory, setKeyAttendaceHistory] = React.useState('his_0');
  const [keySummary, setKeySummary] = React.useState('sum_0');
  const [dataBus, setDataBus] = React.useState([]);
  const [dataLoc, setDataLoc] = React.useState([]);
  const [location, setLocation] = React.useState({} as any);
  const [isLocGranted, setIsLocGranted] = React.useState(false);
  const [loadingSubmit, setLoadingSubmit] = React.useState(false);
  const [selectedMethod, setSelectedMethod] = React.useState('Check in');
  const [showProgress, setShowProgress] = React.useState(false);
  const [progress, setProgress] = React.useState({
    count: 0,
    text: '',
  });
  const [form, setForm] = React.useState({
    WorkLocationID: '',
    Ref1: '',
    Photo: '',
    ImageUri: '',
    File: {} as any,
  });
  const [errorForm, setErrorForm] = React.useState({
    WorkLocationID: false,
    Photo: false,
  });
  //modal
  const [showModal, setShowModal] = React.useState(false);
  const init = () => {
    setKeyAttendaceHistory('his_' + Math.random());
    setKeySummary('sum_' + Math.random());

    getStatus({});
    Promise.all([
      getMasterBus({
        Select: ['_id', 'Name'],
        Sort: ['_id'],
        Take: 25,
      }),
      // getNearestLocation({
      //   Latitude: location?.coords?.latitude,
      //   Longitude: location?.coords?.longitude,
      // }),
    ]).then(responses => {
      const [resBus] = responses;
      setDataBus(resBus);
      // setDataLoc(resLoc);
    });
  };
  const fetchNearestLocation = () => {
    getNearestLocation({
      Latitude: location?.coords?.latitude,
      Longitude: location?.coords?.longitude,
    }).then((resLoc) => {
      setDataLoc(resLoc);
    })
  }
  React.useLayoutEffect(() => {
    ctx.setRefreshCallback({
      func: async () => {
        await getLocation();
        init();
      },
    });
    return () => {};
  }, [isFocused, props.navigation]);

  React.useEffect(() => {
    if (isFocused) {
      getLocation();
      init();
    }
    return () => {};
  }, [isFocused]);

  React.useEffect(() => {
    setKeyStatusBar(Math.random);
    return () => {};
  }, [isCheckIn]);

  const onSubmit = () => {
    if (!isLocGranted) {
      return Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody:
          'Oops! It seems like there was a hiccup in fetching your location. To continue, please grant permission to access your current location.',
      });
    }
    if (form.WorkLocationID === '') {
      setErrorForm({...errorForm, WorkLocationID: true});
      return Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody: 'Location can not be empty',
      });
    }
    if (form.ImageUri === '') {
      setErrorForm({...errorForm, Photo: true});
      return Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody: 'Photo can not be empty',
      });
    }
    setLoadingSubmit(true);
    setShowProgress(true);
    setProgress({
      count: 1,
      text: 'Saving your attendance data…',
    });
    if (selectedMethod === 'Check In') {
      checkIn({
        payload: {
          ...form,
          Op: 'Checkin',
          Time: null,
          Lat: location?.coords?.latitude,
          Long: location?.coords?.longitude,
          ConfirmForReview: false,
        },
      })
        .then(async r => {
          const file = {
            ...form.File,
            Asset: {
              ...form.File?.Asset,
              RefID: r,
            },
          };
          setProgress({
            count: 2,
            text: 'Uploading Process..',
          });
          await writeBatchWithContent([file]).then(() => {
            Toast.show({
              type: ALERT_TYPE.SUCCESS,
              title: 'Success',
              textBody: 'You have successfully checked in. Enjoy your time!',
            });
            setForm({
              WorkLocationID: '',
              Ref1: '',
              Photo: '',
              ImageUri: '',
              File: {},
            });
          }).catch(async (e) => {
            //delete check here
            await deleteAttendance({_id: r})

            return Toast.show({
              type: ALERT_TYPE.DANGER,
              title: 'Error',
              textBody:
                'Oops! Check in is failed, Please try again!',
            });
          });
        })
        .catch(e => {
          Toast.show({
            type: ALERT_TYPE.DANGER,
            title: 'Error',
            textBody: 'Opps! ' + e,
          });
        })
        .finally(() => {
          setLoadingSubmit(false);
          setShowModal(false);
          setShowProgress(false);
          setProgress({
            count: 0,
            text: '',
          });
          setForm({
            WorkLocationID: '',
            Ref1: '',
            Photo: '',
            ImageUri: '',
            File: {},
          });
          init();
        });
    } else {
      checkOut({
        payload: {
          ...form,
          Op: 'Checkout',
          Time: null,
          Lat: location?.coords?.latitude,
          Long: location?.coords?.longitude,
          ConfirmForReview: false,
        },
      })
        .then(async r => {
          const file = {
            ...form.File,
            Asset: {
              ...form.File?.Asset,
              RefID: r,
            },
          };
          setProgress({
            count: 2,
            text: 'Uploading Process..',
          });
          await writeBatchWithContent([file]).then(() => {
            Toast.show({
              type: ALERT_TYPE.SUCCESS,
              title: 'Success',
              textBody: 'You have successfully checked out. Enjoy your time!',
            });
            setForm({
              WorkLocationID: '',
              Ref1: '',
              Photo: '',
              ImageUri: '',
              File: {},
            });
          }).catch(async (e) => {
            //delete check here
            await deleteAttendance({_id: r})
            return Toast.show({
              type: ALERT_TYPE.DANGER,
              title: 'Error',
              textBody:
                'Oops! Check out is failed, Please try again!',
            });
          });
        })
        .catch(e => {
          Toast.show({
            type: ALERT_TYPE.DANGER,
            title: 'Error',
            textBody: 'Opps! ' + e,
          });
        })
        .finally(() => {
          setLoadingSubmit(false);
          setShowModal(false);
          setShowProgress(false);
          setProgress({
            count: 0,
            text: '',
          });
          setForm({
            WorkLocationID: '',
            Ref1: '',
            Photo: '',
            ImageUri: '',
            File: {},
          });
          init();
        });
    }
  };

  // BUS
  const onSearchBus = (text: string) => {
    getMasterBus({
      Select: ['_id', 'Name'],
      Sort: ['_id'],
      Where: {
        Op: '$or',
        items: [
          {
            Field: '_id',
            Op: '$contains',
            Value: [text],
          },
          {
            Field: 'Name',
            Op: '$contains',
            Value: [text],
          },
        ],
      },
    }).then(r => {
      setDataBus(r);
    });
  };
  const onTakePicture = async () => {
    try {
      setErrorForm({...errorForm, Photo: false});
      launchCamera(
        {
          mediaType: 'photo',
          cameraType: 'front',
          saveToPhotos: true,
          includeBase64: true,
          quality: 0.5,
        },
        async (response: any) => {
          if (!response.didCancel === true) {
            const file = response.assets[0];
            const _asset = {
              Asset: {
                ContentType: file.type,
                Data: {
                  UploadDate: new Date(),
                  Content: '',
                  FileName: file.fileName,
                  Description: 'kara attendance',
                },
                Kind: 'Attendance',
                RefID: '',
                OriginalFileName: file.fileName,
              },
              Content: file.base64,
            };
            setForm({...form, ImageUri: file.uri, File: _asset});
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
  //Geo Loc

  const requestLocationPermission = async () => {
    try {
      const granted = await PermissionsAndroid.request(
        PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION,
        {
          title: 'Geolocation Permission',
          message: 'Can we access your location?',
          buttonNeutral: 'Ask Me Later',
          buttonNegative: 'Cancel',
          buttonPositive: 'OK',
        },
      );
      if (granted === 'granted') {
        // console.log('You can use Geolocation');
        return true;
      } else {
        // console.log('You cannot use Geolocation');
        return false;
      }
    } catch (err) {
      return false;
    }
  };
  const getLocation = async () => {
    const result = requestLocationPermission();
    result.then(res => {
      if (res) {
        Geolocation.getCurrentPosition(
          position => {
            // setLocation(position);
            setLocation(position);
            setIsLocGranted(true);
          },
          error => {
            // See error code charts below.
            console.log(error.code, error.message);
            setLocation(false);
          },
          {enableHighAccuracy: true, timeout: 15000, maximumAge: 10000},
        );
      }
    });
  };
  // // const MINUTE_MS = 10000;
  // React.useEffect(() => {
  //   // const interval = setInterval(() => {
  //   //   getLocation();
  //   // }, MINUTE_MS);
  //   getLocation();
  //   // return () => clearInterval(interval); // This represents the unmount function, in which you need to clear your interval to prevent memory leaks.
  //   return () => {}; // This represents the unmount function, in which you need to clear your interval to prevent memory leaks.
  // }, []);
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
    <>
      {isFocused && (
        <>
          <StatusBar
            key={keyStatusBar}
            barStyle="dark-content"
            backgroundColor={Colors.SHADES.gray[100]}
          />
          <View
            style={{
              ...styles.container,
              backgroundColor: Colors.SHADES.gray[100],
            }}>
            <DateLocInformation location={location} />
            <CardAttendanceInfo key={keySummary} />
            <AttendanceHistory
              key={keyAttendaceHistory}
              navigation={props.navigation}
              isFocused={isFocused}
            />
            {!isCheckIn ? (
              <Button
                onPress={() => {
                  fetchNearestLocation()
                  setForm({
                    WorkLocationID: '',
                    Ref1: '',
                    Photo: '',
                    ImageUri: '',
                    File: {},
                  });
                  setSelectedMethod('Check In');
                  setShowModal(true);
                }}
                style={styles.checkinButton}
                size="large"
                status="primary"
                accessoryLeft={() => {
                  return (
                    <Image
                      style={{
                        width: Mixins.scaleSize(40),
                        height: Mixins.scaleSize(40),
                      }}
                      source={images.iconFinger}
                    />
                  );
                }}>
                {() => (
                  <Text style={{...Typography.textLgSemiBold, color: 'white'}}>
                    Tap to Check In
                  </Text>
                )}
              </Button>
            ) : (
              <Button
                onPress={() => {
                  // checkOut({
                  //   payload: {
                  //     Op: 'Checkout',
                  //     WorkLocationID: dataStatus.WorkLocationID,
                  //     Lat: location?.coords?.latitude,
                  //     Long: location?.coords?.longitude,
                  //   },
                  // });
                  fetchNearestLocation()
                  setForm({
                    WorkLocationID: '',
                    Ref1: '',
                    Photo: '',
                    ImageUri: '',
                    File: {},
                  });
                  setSelectedMethod('Check Out');
                  setShowModal(true);
                }}
                style={styles.checkinButton}
                size="large"
                status="warning"
                accessoryLeft={() => {
                  return (
                    <Image
                      style={{
                        width: Mixins.scaleSize(40),
                        height: Mixins.scaleSize(40),
                      }}
                      source={images.iconFinger}
                    />
                  );
                }}>
                {() => (
                  <Text style={{...Typography.textLgSemiBold, color: 'white'}}>
                    Tap to Check Out
                  </Text>
                )}
              </Button>
            )}
          </View>
          <Modal
            isVisible={showModal}
            onBackdropPress={() => setShowModal(false)}
            backdropColor="#B4B3DB"
            backdropOpacity={0.8}
            animationIn="zoomInDown"
            animationOut="zoomOutUp"
            animationInTiming={600}
            animationOutTiming={600}
            backdropTransitionInTiming={600}
            backdropTransitionOutTiming={600}>
            <View style={styles.modalContainer}>
              <View style={styles.modalHeader}>
                <Text style={styles.modalLabel}>{selectedMethod} Form</Text>
                <TouchableOpacity onPress={() => setShowModal(false)}>
                  <Icon
                    name="close-circle-outline"
                    style={{width: 30, height: 30}}
                    fill={Colors.SHADES.dark[50]}
                  />
                </TouchableOpacity>
              </View>
              <View style={styles.modalBody}>
                <ScrollView>
                  <Dropdown
                    items={dataLoc.map((o: any) => {
                      return {value: o._id, label: o.Name};
                    })}
                    value={form.WorkLocationID}
                    onSelect={(selected: any) => {
                      setForm({...form, WorkLocationID: selected});
                      setErrorForm({...errorForm, WorkLocationID: false});
                    }}
                    inputLabel="Location"
                    isRequired
                    error={errorForm.WorkLocationID}
                  />
                  {selectedMethod === 'Check In' && (
                    <Dropdown
                      items={dataBus.map((o: any) => {
                        return {value: o._id, label: o.Name};
                      })}
                      value={form.Ref1}
                      onSelect={(selected: any) => {
                        setForm({...form, Ref1: selected});
                      }}
                      inputLabel="Bus"
                      onSearch={onSearchBus}
                    />
                  )}
                  <View
                    style={{
                      marginTop: Mixins.scaleSize(10),
                      marginBottom: Mixins.scaleSize(20),
                    }}>
                    <View style={{alignContent: 'center'}}>
                      <TouchableOpacity
                        onPress={() => onTakePicture()}
                        style={styles.boxFile}>
                        {form.ImageUri ? (
                          <Image
                            source={{uri: form.ImageUri}}
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
                        {form.ImageUri && (
                          <TouchableOpacity
                            style={styles.closebutton}
                            onPress={() => {
                              setForm({...form, ImageUri: ''});
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
                </ScrollView>
              </View>
              <View style={styles.modalFooter}>
                <Button
                  disabled={loadingSubmit}
                  size="large"
                  style={styles.buttonSubmit}
                  status="primary"
                  // accessoryLeft={
                  //   loadingSubmit ? (
                  //     <View
                  //       style={{
                  //         alignItems: 'center',
                  //       }}>
                  //       <Spinner size="small" />
                  //     </View>
                  //   ) : undefined
                  // }
                  onPress={onSubmit}>
                  <Text style={styles.buttonLabel}>Submit</Text>
                </Button>
              </View>
            </View>
          </Modal>
          <Modal
            isVisible={showProgress}
            backdropColor="#B4B3DB"
            backdropOpacity={0.8}
            animationIn="zoomInDown"
            animationOut="zoomOutUp"
            animationInTiming={600}
            animationOutTiming={600}
            backdropTransitionInTiming={600}
            backdropTransitionOutTiming={600}>
            <View style={styles.modalProgressContainer}>
              {/* <View style={{ backgroundColor: Colors.WHITE, opacity: 0.9}}> */}
              {/* <Progress.Circle
                size={Mixins.scaleSize(100)}
                borderWidth={4}
                borderColor={theme['color-primary-500']}
                color={theme['color-primary-600']}
                strokeCap="round"
                indeterminate={true}
                endAngle={0.7}
                progress={0.2}
                showsText
              /> */}
              <Progress.Bar
                width={Mixins.scaleSize(150)}
                height={Mixins.scaleSize(15)}
                borderRadius={10}
                color={theme['color-primary-600']}
                animationType="spring"
                indeterminate={true}
              />
              <View style={{marginTop: Mixins.scaleSize(20)}}>
                <Animated.Text
                  style={{
                    ...Typography.textLgPlusSemiBold,
                    transform: [{scale: scaleValue}],
                  }}>
                  {progress.text}
                </Animated.Text>
              </View>
              {/* </View> */}
            </View>
          </Modal>
        </>
      )}
    </>
  );
};

export default container(Layout, true);

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingVertical: Mixins.scaleSize(10),
    paddingHorizontal: Mixins.scaleSize(14),
  },
  icon: {
    width: Mixins.scaleSize(90),
    height: Mixins.scaleSize(90),
  },
  button: {
    borderRadius: 100,
    backgroundColor: theme['color-info-100'],
    alignItems: 'center',
    justifyContent: 'center',
    width: 200,
    height: 200,
    borderWidth: 20,
  },
  buttonLabel: {
    ...Typography.textLgPlusSemiBold,
  },
  card: {
    // width: '100%',
    marginHorizontal: Mixins.scaleSize(20),
    borderRadius: 10,
    borderColor: Colors.SHADES.dark[100],
    shadowColor: Colors.SHADES.dark[100],
    shadowOffset: {width: 0, height: 0},
    shadowOpacity: 70,
    opacity: 0.7,
    backgroundColor: Colors.WHITE,
  },
  modalContainer: {
    flex: 1,
    backgroundColor: Colors.WHITE,
    borderRadius: Mixins.scaleSize(15),
    padding: Mixins.scaleSize(10),
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: Mixins.scaleSize(10),
    alignItems: 'center',
    padding: Mixins.scaleSize(10),
  },
  modalLabel: {
    ...Typography.textLgPlusSemiBold,
    color: Colors.BLACK,
  },
  modalBody: {
    flex: 1,
    padding: Mixins.scaleSize(10),
  },
  modalFooter: {
    padding: Mixins.scaleSize(10),
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
    height: Mixins.scaleSize(200),
    position: 'relative',
  },
  closebutton: {
    position: 'absolute',
    zIndex: 10,
    right: Mixins.scaleSize(-2),
    backgroundColor: Colors.WHITE,
    borderRadius: 50,
    top: Mixins.scaleSize(-10),
  },
  checkinButton: {
    borderRadius: Mixins.scaleSize(8),
    // backgroundColor: Colors.SHADES.red[400],
    alignItems: 'center',
    justifyContent: 'center',
  },
  buttonSubmit: {
    borderRadius: Mixins.scaleSize(8),
    alignItems: 'center',
    justifyContent: 'center',
  },
  modalProgressContainer: {
    flex: 1,
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
  },
});
