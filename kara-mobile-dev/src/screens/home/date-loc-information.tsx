/* eslint-disable react-hooks/exhaustive-deps */
import {StyleSheet, Text, View} from 'react-native';
import React from 'react';
import moment from 'moment';
import {Colors, Mixins, Typography} from 'utils';
import {Icon} from '@ui-kitten/components';

type Props = {
  location: any;
};

const Layout = (props: Props) => {
  const {location} = props;
  const [detailLoc, setDetailLoc] = React.useState(null as any);
  const [dt, setDt] = React.useState(new Date());

  const getLoc = async () => {
    const url = `https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=${location?.coords?.latitude}&lon=${location?.coords?.longitude}`;
    const response = await fetch(url);
    const data = await response.json();
    setDetailLoc(data);
    // console.log('asasd', url, data);
  };
  React.useEffect(() => {
    let secTimer = setInterval(() => {
      setDt(new Date());
    }, 1000);

    return () => clearInterval(secTimer);
  }, []);
  React.useEffect(() => {
    // console.log(location);
    getLoc();
    return () => {};
  }, [location]);

  return (
    <View style={styles.container}>
      <View>
        <Text style={{...Typography.textMd, color: Colors.BLACK}}>
          {moment(dt).local().format('ddd, DD MMM YYYY')}
        </Text>
        <Text style={{...Typography.textMd, color: Colors.BLACK}}>
          {moment(dt).local().format('HH:mm')}
        </Text>
      </View>
      <View style={styles.location}>
        <Icon style={styles.icon} fill="white" name="pin-outline" />
        {detailLoc !== null && (
          <Text style={{...Typography.textMd, color: Colors.WHITE}}>
          {detailLoc?.address?.city ?? detailLoc?.address?.county}, {detailLoc?.address?.state}, {detailLoc?.address?.country_code.toUpperCase()}
        </Text>
        )}
      </View>
    </View>
  );
};

export default Layout;

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: Mixins.scaleSize(20),
  },
  location: {
    flexDirection: 'row',
    gap: 5,
    paddingHorizontal: Mixins.scaleSize(15),
    alignItems: 'center',
    backgroundColor: Colors.SHADES.red[400],
    color: Colors.WHITE,
    borderRadius: 50,
  },
  icon: {
    width: Mixins.scaleSize(15),
    height: Mixins.scaleSize(15),
  },
});
