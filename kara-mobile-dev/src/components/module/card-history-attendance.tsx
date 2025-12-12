/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react-native/no-inline-styles */
import {StyleSheet, Text, TouchableOpacity, View} from 'react-native';
import React from 'react';
import {Colors, Helper, Mixins, Typography} from 'utils';
import {Icon, Spinner} from '@ui-kitten/components';
import moment from 'moment';
import {useActions} from '@overmind/index';
import ImageView from 'react-native-image-viewing';
import {ALERT_TYPE, Toast} from 'react-native-alert-notification';

type Props = {
  data: any;
};

const Layout = (props: Props) => {
  const {data} = props;
  const {getsAssetByJournal} = useActions();
  const [loadingShowPhoto, setLoadingShowPhoto] = React.useState(false);
  const [show, setShow] = React.useState(false);
  const [images, setImages] = React.useState([] as any);

  const decimalToHoursMinutes = (value: any) => {
    var hours = Math.floor(value);
    var minutes = Math.round((value - hours) * 60);
    return hours + ' : ' + minutes;
  };
  const onShowImage = (item: any) => {
    setLoadingShowPhoto(true);
    getsAssetByJournal({
      JournalType: 'Attendance',
      JournalID: item,
    })
      .then((res: any) => {
        const file = res[0];
        if (file) {
          const img = Helper.getAssetURL(file._id);
          setImages([
            {
              uri: img,
            },
          ]);
          setShow(true);
        }
        
        setLoadingShowPhoto(false);
      })
      .catch(e => {
        setLoadingShowPhoto(false);
        return Toast.show({
          type: ALERT_TYPE.DANGER,
          title: 'Error',
          textBody: e,
        });
      });
    // getPhoto(item)
    //   .then(r => {
    //     const img = 'data:image/jpeg;base64, ' + r.Photo;
    //     setImages([
    //       {
    //         uri: img,
    //       },
    //     ]);
    //     setShow(true);
    //     setLoadingShowPhoto(false);
    //   })
    //   .catch(e => {
    //     setLoadingShowPhoto(false);
    //     return Toast.show({
    //       type: ALERT_TYPE.DANGER,
    //       title: 'Error',
    //       textBody: e,
    //     });
    //   });
  };
  return (
    <>
      <View style={styles.card}>
        <View style={styles.section}>
          <View style={styles.dateContainer}>
            <Text style={styles.dateLabel}>
              {moment(data.CheckIn).format('DD')}
            </Text>
            <Text style={styles.monthLabel}>
              {moment(data.CheckIn).format('MMM')}
            </Text>
          </View>

          <View style={{flexDirection: 'column', justifyContent: 'center'}}>
            <View style={styles.dateSection}>
              <View style={{...styles.date, borderRightWidth: 1}}>
                <Text
                  style={{
                    ...Typography.textMdPlus,
                    color: Colors.BLACK,
                    textAlign: 'center',
                  }}>
                  {moment(data.CheckIn).format('hh : mm')}
                </Text>
                <Text
                  style={{
                    ...Typography.textMd,
                    textAlign: 'center',
                    color: Colors.BLACK,
                  }}>
                  Check In
                </Text>
                {data.CheckIn !== null && (
                  <TouchableOpacity
                    onPress={() => onShowImage(data.CheckInID)}
                    disabled={loadingShowPhoto}
                    style={{
                      borderRadius: Mixins.scaleSize(12),
                      backgroundColor: Colors.WHITE,
                      paddingHorizontal: Mixins.scaleSize(10),
                      paddingVertical: Mixins.scaleSize(5),
                    }}>
                    {loadingShowPhoto ? (
                      <Spinner />
                    ) : (
                      <Icon
                        style={styles.iconCamera}
                        fill={Colors.SHADES.green['500']}
                        name="eye-outline"
                      />
                    )}
                  </TouchableOpacity>
                )}
              </View>
              <View style={{...styles.date, borderRightWidth: 1}}>
                <Text
                  style={{
                    ...Typography.textMdPlus,
                    color: Colors.BLACK,
                    textAlign: 'center',
                  }}>
                  {data.CheckOut !== null
                    ? moment(data.CheckOut).format('hh : mm')
                    : '-'}
                </Text>
                <Text
                  style={{
                    ...Typography.textMd,
                    textAlign: 'center',
                    color: Colors.BLACK,
                  }}>
                  Check Out
                </Text>
                {data.CheckOut !== null && (
                  <TouchableOpacity
                    onPress={() => onShowImage(data.CheckOutID)}
                    disabled={loadingShowPhoto}
                    style={{
                      borderRadius: Mixins.scaleSize(12),
                      backgroundColor: Colors.WHITE,
                      paddingHorizontal: Mixins.scaleSize(10),
                      paddingVertical: Mixins.scaleSize(5),
                    }}>
                    {loadingShowPhoto ? (
                      <Spinner />
                    ) : (
                      <Icon
                        style={styles.iconCamera}
                        fill={Colors.SHADES.green['500']}
                        name="eye-outline"
                      />
                    )}
                  </TouchableOpacity>
                )}
              </View>
              <View style={styles.date}>
                <Text
                  style={{
                    ...Typography.textMdPlus,
                    color: Colors.BLACK,
                    textAlign: 'center',
                  }}>
                  {decimalToHoursMinutes(data.TotalHour)}
                </Text>
                <Text
                  style={{
                    ...Typography.textMd,
                    textAlign: 'center',
                    color: Colors.BLACK,
                  }}>
                  Total Hours
                </Text>
              </View>
            </View>
            <View style={styles.locSection}>
              <Icon
                style={styles.icon}
                fill={Colors.SHADES.red[600]}
                name="pin-outline"
              />
              <Text style={{...Typography.textMdPlus, color: Colors.BLACK}}>
                {data.WorkLocationName}
              </Text>
            </View>
          </View>
        </View>
      </View>
      <ImageView
        images={images}
        imageIndex={0}
        visible={show}
        onRequestClose={() => setShow(false)}
      />
    </>
  );
};

export default Layout;

const styles = StyleSheet.create({
  card: {
    marginBottom: Mixins.scaleSize(10),
    padding: Mixins.scaleSize(12),
    backgroundColor: Colors.WHITE,
    borderRadius: Mixins.scaleSize(12),
    borderWidth: 1,
    borderColor: Colors.SHADES.gray[100],
  },
  section: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Mixins.scaleSize(10),
  },
  dateContainer: {
    height: '100%',
    flexDirection: 'column',
    alignItems: 'center',
    backgroundColor: Colors.SHADES.gray[500],
    padding: Mixins.scaleSize(10),
    paddingVertical: Mixins.scaleSize(25),
    borderRadius: Mixins.scaleSize(12),
  },
  dateLabel: {
    ...Typography.textLgSemiBold,
    textAlign: 'center',
    color: Colors.WHITE,
  },
  monthLabel: {
    ...Typography.textLg,
    textAlign: 'center',
    color: Colors.WHITE,
  },
  dateSection: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: Mixins.scaleSize(10),
    marginBottom: Mixins.scaleSize(10),
  },
  date: {
    paddingHorizontal: Mixins.scaleSize(10),
  },
  locSection: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    gap: Mixins.scaleSize(10),
  },
  icon: {
    width: Mixins.scaleSize(15),
    height: Mixins.scaleSize(15),
  },
  iconCamera: {
    width: Mixins.scaleSize(17),
    height: Mixins.scaleSize(17),
  },
});
