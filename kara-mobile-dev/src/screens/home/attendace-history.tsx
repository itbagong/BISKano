/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable react-native/no-inline-styles */
import {
  ActivityIndicator,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import CardAttendanceHistory from '@components/module/card-history-attendance';
import {Colors, Mixins, Typography} from 'utils';
import s from '@components/styles';
import {default as theme} from '../../custom-theme.json';
import {useActions} from '@overmind/index';

type Props = {
  navigation: any;
  isFocused: any;
};

const Layout = (props: Props) => {
  const {navigation} = props;
  const {getListHistory} = useActions();
  const [loading, setLoading] = React.useState(false);
  const [data, setData] = React.useState([]);

  const init = () => {
    setLoading(true);
    getListHistory({sort: ['-_id']})
      .then(res => {
        setData(res.slice(0, 2));
      })
      .finally(() => {
        setLoading(false);
      });
  };

  React.useEffect(() => {
    init();
  }, []);

  return (
    <>
      <View style={styles.container}>
        {loading ? (
          <>
            <ActivityIndicator color={Colors.PRIMARY.red} />
          </>
        ) : (
          <>
            <View
              style={{
                ...s.row,
                justifyContent: 'space-between',
                alignItems: 'center',
              }}>
              <Text style={styles.title}>Attendance History</Text>
              <TouchableOpacity onPress={() => navigation.navigate('History')}>
                <Text style={styles.label}>See more</Text>
              </TouchableOpacity>
            </View>
            {data?.map((item: any, i: number) => (
              <CardAttendanceHistory key={i} data={item} />
            ))}
          </>
        )}
      </View>
    </>
  );
};

export default Layout;

const styles = StyleSheet.create({
  container: {
    marginBottom: Mixins.scaleSize(20),
  },
  title: {
    color: Colors.BLACK,
    ...Typography.textLgSemiBold,
  },
  label: {
    ...Typography.textMdPlusSemiBold,
    color: theme['color-primary-600'],
  },
});
