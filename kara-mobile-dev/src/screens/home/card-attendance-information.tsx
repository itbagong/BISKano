/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable react-native/no-inline-styles */
import {StyleSheet, Text, View, ActivityIndicator} from 'react-native';
import React from 'react';
import {Colors, Mixins, Typography} from 'utils';
// import {default as theme} from '../../custom-theme.json'; // <-- Import app theme
import {Card, Icon} from '@ui-kitten/components';
import {useActions, useAppState} from '@overmind/index';
import moment from 'moment';

type Props = {};

const Layout = (props: Props) => {
  const {} = props;
  const {getAttendanceSummary} = useActions();
  const {dataStatus} = useAppState();
  const [data, setData] = React.useState({} as any);
  const [loading, setLoading] = React.useState(false);
  React.useEffect(() => {
    setLoading(true);
    getAttendanceSummary({month: moment().format('YYYYMM')})
      .then(res => {
        setData(res);
      })
      .finally(() => setLoading(false));
    return () => {};
  }, []);
  return (
    <View style={styles.container}>
      {loading ? (
        <>
          <ActivityIndicator color={Colors.PRIMARY.red} />
        </>
      ) : (
        <>
          <View style={styles.sectionRow}>
            <Card style={styles.card}>
              <View style={styles.cardHeader}>
                <View>
                  <View style={styles.iconContainer}>
                    <Icon
                      style={styles.icon}
                      name="diagonal-arrow-left-down"
                      fill="white"
                    />
                  </View>
                </View>
                <View>
                  <Text
                    style={{...Typography.textMdPlusSemiBold, color: 'black'}}>
                    Check in
                  </Text>
                  <Text style={{...Typography.textMd, color: Colors.BLACK}}>
                    {dataStatus?.CheckinTime !== null ? 'Checked' : 'Not yet'}
                  </Text>
                </View>
              </View>
              <View>
                <Text
                  style={{...Typography.headerLgSemiBold, color: Colors.BLACK}}>
                  {dataStatus?.CheckinTime !== null
                    ? moment(dataStatus?.CheckinTime).format('HH:mm')
                    : '00:00'}
                </Text>
              </View>
            </Card>
            <Card style={styles.card}>
              <View style={styles.cardHeader}>
                <View>
                  <View style={styles.iconContainer}>
                    <Icon
                      style={styles.icon}
                      name="diagonal-arrow-right-up"
                      fill="white"
                    />
                  </View>
                </View>
                <View>
                  <Text
                    style={{...Typography.textMdPlusSemiBold, color: 'black'}}>
                    Check out
                  </Text>
                  <Text style={{...Typography.textMd, color: Colors.BLACK}}>
                    {dataStatus?.CheckoutTime !== null ? 'Checked' : 'Not yet'}
                  </Text>
                </View>
              </View>
              <View>
                <Text
                  style={{...Typography.headerLgSemiBold, color: Colors.BLACK}}>
                  {dataStatus?.CheckoutTime !== null
                    ? moment(dataStatus?.CheckoutTime).format('HH:mm')
                    : '00:00'}
                </Text>
              </View>
            </Card>
          </View>
          <View style={styles.sectionRow}>
            <Card style={styles.card}>
              <View style={styles.cardHeader}>
                <View>
                  <View style={styles.iconContainer}>
                    <Icon
                      style={styles.icon}
                      name="arrow-upward"
                      fill="white"
                    />
                  </View>
                </View>
                <View>
                  <Text
                    style={{...Typography.textMdPlusSemiBold, color: 'black'}}>
                    Absence
                  </Text>
                  <Text style={{...Typography.textMd, color: Colors.BLACK}}>
                    {moment().format('MMMM YYYY')}
                  </Text>
                </View>
              </View>
              <View
                style={{flexDirection: 'row', gap: 10, alignItems: 'center'}}>
                <Text
                  style={{...Typography.headerLgSemiBold, color: Colors.BLACK}}>
                  {data?.Absence}
                </Text>
                <Text
                  style={{
                    ...Typography.textLg,
                    color: Colors.SHADES.dark[300],
                  }}>
                  day
                </Text>
              </View>
            </Card>
            <Card style={styles.card}>
              <View style={styles.cardHeader}>
                <View>
                  <View style={styles.iconContainer}>
                    <Icon
                      style={styles.icon}
                      name="flip-2-outline"
                      fill="white"
                    />
                  </View>
                </View>
                <View>
                  <Text
                    style={{...Typography.textMdPlusSemiBold, color: 'black'}}>
                    Attended
                  </Text>
                  <Text style={{...Typography.textMd, color: Colors.BLACK}}>
                    {moment().format('MMMM YYYY')}
                  </Text>
                </View>
              </View>
              <View
                style={{flexDirection: 'row', gap: 10, alignItems: 'center'}}>
                <Text
                  style={{...Typography.headerLgSemiBold, color: Colors.BLACK}}>
                  {data?.TotalAttendance}
                </Text>
                <Text
                  style={{
                    ...Typography.textLg,
                    color: Colors.SHADES.dark[300],
                  }}>
                  day
                </Text>
              </View>
            </Card>
          </View>
        </>
      )}
    </View>
  );
};

export default Layout;

const styles = StyleSheet.create({
  container: {
    marginBottom: Mixins.scaleSize(20),
    // paddingHorizontal: Mixins.scaleSize(20),
  },
  sectionRow: {
    flex: 1,
    flexDirection: 'row',
    gap: Mixins.scaleSize(10),
    marginBottom: Mixins.scaleSize(10),
  },
  card: {
    flex: 1,
    borderRadius: Mixins.scaleSize(12),
  },
  cardHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Mixins.scaleSize(10),
    marginBottom: Mixins.scaleSize(10),
  },
  iconContainer: {
    backgroundColor: Colors.SHADES.gray[900],
    borderRadius: 100,
    width: Mixins.scaleSize(40),
    height: Mixins.scaleSize(40),
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
  },
  icon: {
    width: Mixins.scaleSize(25),
    height: Mixins.scaleSize(25),
  },
});
